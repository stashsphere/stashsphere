import logging
import requests
import unittest
from urllib.parse import urlparse, parse_qs
from bs4 import BeautifulSoup

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

BASE = "http://127.0.0.1:8081"
DEX = "http://127.0.0.1:5556"
FRONTEND = "http://localhost:3000"

def dex_login(email, password):
    """
    Drives the full OIDC authorize -> Dex login -> callback chain.
    Returns (session, redirect_url) where redirect_url is the final
    frontend redirect that StashSphere issued.
    """
    s = requests.Session()

    # Start the authorize flow - get redirect to Dex
    r = s.get(f"{BASE}/api/auth/oidc/dex/authorize", allow_redirects=False)
    if r.status_code != 302:
        raise AssertionError(f"authorize redirect: expected 302, got {r.status_code}")
    dex_auth_url = r.headers["Location"]
    if "/dex/auth" not in dex_auth_url:
        raise AssertionError(f"redirect to dex: expected '/dex/auth' in {dex_auth_url!r}")
    logger.debug("authorize -> %s", dex_auth_url)

    # Follow redirect to Dex - this may go through /dex/auth -> /dex/auth/local
    # Let requests handle intermediate Dex redirects automatically
    r = s.get(dex_auth_url)
    if r.status_code != 200:
        raise AssertionError(f"dex login page: expected 200, got {r.status_code}")
    logger.debug("login page URL: %s", r.url)
    logger.debug("login page HTML (first 1000 chars):\n%s", r.text[:1000])

    # Parse the login page with BeautifulSoup
    soup = BeautifulSoup(r.text, "html.parser")

    # Extract form action
    form = soup.find("form")
    if not form:
        raise AssertionError("could not find form in Dex HTML")

    form_action = form.get("action", "")
    # Resolve relative URLs against the current page URL
    if form_action.startswith("/"):
        form_action = f"{DEX}{form_action}"
    elif not form_action.startswith("http"):
        # Relative to current URL
        base = r.url.rsplit("/", 1)[0]
        form_action = f"{base}/{form_action}"
    logger.debug("form action: %s", form_action)

    # Extract all hidden input fields
    form_data = {"login": email, "password": password}
    for input_tag in form.find_all("input", type="hidden"):
        name = input_tag.get("name")
        value = input_tag.get("value", "")
        if name:
            form_data[name] = value
    logger.debug("form data keys: %s", list(form_data.keys()))

    # Submit login form to Dex
    r = s.post(form_action, data=form_data, allow_redirects=False)
    logger.debug("POST login -> status=%s location=%s", r.status_code, r.headers.get('Location', 'none'))

    # Follow redirects manually — stop when we reach the frontend URL
    max_redirects = 10
    for i in range(max_redirects):
        if r.status_code not in (301, 302, 303, 307, 308):
            break
        location = r.headers.get("Location", "")
        logger.debug("redirect[%s] → %s", i, location)
        # If this is the final redirect to the frontend, stop here
        if location.startswith(FRONTEND):
            return s, location
        r = s.get(location, allow_redirects=False)

    if r.status_code in (301, 302, 303, 307, 308):
        return s, r.headers.get("Location", "")

    raise AssertionError(f"did not reach frontend redirect after login, last status={r.status_code}, url={getattr(r, 'url', '?')}, body:\n{r.text[:1000]}")


class TestOIDC(unittest.TestCase):
    alice_id = None

    # Info endpoint includes Dex provider
    def test_01_info(self):
        logger.info("=== test_info ===")
        r = requests.get(f"{BASE}/api/info")
        self.assertEqual(r.status_code, 200, "info status")
        body = r.json()
        providers = body.get("oidcProviders", [])
        self.assertEqual(len(providers), 1, "provider count")
        self.assertEqual(providers[0]["name"], "dex", "provider name")
        self.assertEqual(providers[0]["displayName"], "Dex", "provider displayName")
        logger.info("OK")

    # Full OIDC login - new user
    def test_02_oidc_login_new_user(self):
        logger.info("=== test_oidc_login_new_user ===")
        s, redirect_url = dex_login("alice@example.com", "password")

        self.assertIn("status=success", redirect_url, "callback success redirect")

        # Verify auth cookies are set
        cookie_names = {c.name for c in s.cookies}
        self.assertIn("stashsphere-access", cookie_names, "access cookie")
        self.assertIn("stashsphere-info", cookie_names, "info cookie")

        # Verify profile returns alice's email
        r = s.get(f"{BASE}/api/user/profile")
        self.assertEqual(r.status_code, 200, "profile status")
        profile = r.json()
        self.assertEqual(profile["email"], "alice@example.com", "profile email")
        logger.info("OK (userId=%s)", profile['id'])

        # Store alice_id for the next test
        TestOIDC.alice_id = profile["id"]

    # Full OIDC login - returning user
    def test_03_oidc_login_returning_user(self):
        logger.info("=== test_oidc_login_returning_user ===")
        self.assertIsNotNone(TestOIDC.alice_id, "alice_id must be set from previous test")

        s, redirect_url = dex_login("alice@example.com", "password")
        self.assertIn("status=success", redirect_url, "callback success redirect")

        r = s.get(f"{BASE}/api/user/profile")
        self.assertEqual(r.status_code, 200, "profile status")
        profile = r.json()
        self.assertEqual(profile["id"], TestOIDC.alice_id, "same user id")
        logger.info("OK")

    # Account linking - existing password user
    def test_04_account_linking(self):
        logger.info("=== test_account_linking ===")

        # Register bob as a password user first
        r = requests.post(f"{BASE}/api/user/register", json={
            "name": "Bob",
            "email": "bob@example.com",
            "password": "bob-password",
            "inviteCode": "1234",
        })
        self.assertEqual(r.status_code, 200, "register bob")

        s, redirect_url = dex_login("bob@example.com", "password")

        parsed = urlparse(redirect_url)
        params = parse_qs(parsed.query)
        self.assertEqual(params.get("action", [None])[0], "link_required", "link_required action")
        self.assertEqual(params.get("email", [None])[0], "bob@example.com", "link email")

        challenge_token = None
        for c in s.cookies:
            if c.name == "oidc-link-challenge":
                challenge_token = c.value
                break
        self.assertIsNotNone(challenge_token, "oidc-link-challenge cookie not set")

        r = s.post(f"{BASE}/api/auth/oidc/dex/link", json={
            "password": "bob-password",
            "challengeToken": challenge_token,
        })
        self.assertEqual(r.status_code, 200, "link status")

        cookie_names = {c.name for c in s.cookies}
        self.assertIn("stashsphere-access", cookie_names, "access cookie after link")

        r = s.get(f"{BASE}/api/user/profile")
        self.assertEqual(r.status_code, 200, "profile status")
        profile = r.json()
        self.assertEqual(profile["email"], "bob@example.com", "profile email")

        s2 = requests.Session()
        r = s2.post(f"{BASE}/api/user/login", json={
            "email": "bob@example.com",
            "password": "bob-password",
        })
        self.assertEqual(r.status_code, 200, "password login after link")
        logger.info("OK")

    # Account linking - wrong password
    def test_05_account_linking_wrong_password(self):
        logger.info("=== test_account_linking_wrong_password ===")

        # Register charlie as a password user
        r = requests.post(f"{BASE}/api/user/register", json={
            "name": "Charlie",
            "email": "charlie@example.com",
            "password": "charlie-password",
            "inviteCode": "1234",
        })
        self.assertEqual(r.status_code, 200, "register charlie")

        # Start OIDC login with charlie's Dex account
        s, redirect_url = dex_login("charlie@example.com", "password")

        parsed = urlparse(redirect_url)
        params = parse_qs(parsed.query)
        self.assertEqual(params.get("action", [None])[0], "link_required", "link_required action")
        self.assertEqual(params.get("email", [None])[0], "charlie@example.com", "link email")

        challenge_token = None
        for c in s.cookies:
            if c.name == "oidc-link-challenge":
                challenge_token = c.value
                break
        self.assertIsNotNone(challenge_token, "oidc-link-challenge cookie not set")

        r = s.post(f"{BASE}/api/auth/oidc/dex/link", json={
            "password": "wrong-password",
            "challengeToken": challenge_token,
        })
        self.assertEqual(r.status_code, 401, "link with wrong password should fail")

        cookie_names = {c.name for c in s.cookies}
        self.assertNotIn("stashsphere-access", cookie_names, "access cookie should not be set after failed link")

        logger.info("OK")


if __name__ == "__main__":
    unittest.main()
