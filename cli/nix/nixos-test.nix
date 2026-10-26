{ testers, pkgs }:

testers.nixosTest {
  name = "stashctl";

  nodes.server =
    { ... }:
    {
      imports = [ ../../backend/nix/module.nix ];
      services.stashsphere =
        let
          secretConfig = pkgs.writeText "secret.json" (
            builtins.toJSON {
              invites = {
                enabled = true;
                code = "1234";
              };
              # the test talks plain HTTP to the server
              auth.disableSecureCookies = true;
            }
          );
        in
        {
          enable = true;
          settings = {
            database = {
              host = "/run/postgresql";
              password = "foo";
            };
          };
          configFiles = [ "${secretConfig}" ];
          usesLocalPostgresql = true;
        };
      services.postgresql = {
        enable = true;
        ensureDatabases = [ "stashsphere" ];
        ensureUsers = [
          {
            name = "stashsphere";
            ensureDBOwnership = true;
          }
        ];
      };
      environment.systemPackages = [ pkgs.stashctl ];
      environment.variables.STASHCTL_HOME = "/tmp/stashctl";
    };

  testScript = ''
    import json
    import time

    server_url = "http://127.0.0.1:8081"
    cookies = "/tmp/stashctl/cookies.json"

    start_all()
    server.wait_for_unit("stashsphere.service")
    server.wait_until_succeeds(f"${pkgs.curl}/bin/curl -f {server_url}/api/info")

    # login, refresh and register share a rate limit of a burst of 5
    # requests per IP, refilled by one every 12 seconds, so the test logs
    # in with curl as little as possible

    def api(method, path, body=None):
        """Calls the API with curl as alice, with the access token stashctl stored."""
        stored = json.loads(server.succeed(f"cat {cookies}"))
        access = next(
            cookie["raw_cookie"].split(";")[0] for cookie in stored
            if cookie["raw_cookie"].startswith("stashsphere-access=")
        )
        command = (
            f"${pkgs.curl}/bin/curl -sf -X {method} -H 'cookie: {access}' "
            + f"-H 'content-type: application/json' {server_url}{path}"
        )
        if body is not None:
            command += f" -d '{json.dumps(body)}'"
        out = server.succeed(command)
        return json.loads(out) if out.strip() else None

    def register(user):
        body = json.dumps({
            "name": user,
            "email": f"{user}@example.com",
            "password": "secret",
            "inviteCode": "1234",
        })
        server.succeed(
            "${pkgs.curl}/bin/curl -sf -H 'content-type: application/json' "
            + f"-d '{body}' {server_url}/api/user/register"
        )

    def stashctl_json(command, home="/tmp/stashctl"):
        return json.loads(server.succeed(f"STASHCTL_HOME={home} stashctl --json {command}"))

    register("alice")
    register("bob")

    with subtest("commands fail with exit code 3 before login"):
        status, out = server.execute("stashctl --json whoami 2>&1")
        assert status == 3, f"exit code {status}: {out}"
        assert "not logged in" in json.loads(out)["error"]["message"], out

    with subtest("a wrong password is rejected"):
        status, out = server.execute(
            f"echo wrong | stashctl --server {server_url} login --email alice@example.com --password-stdin 2>&1"
        )
        assert status == 3, f"exit code {status}: {out}"
        assert out.strip() == "error: Unauthorized (HTTP 401)", out

    with subtest("login stores the session"):
        server.succeed(
            f"echo secret | stashctl --server {server_url} login --email alice@example.com --password-stdin"
        )
        assert server.succeed(f"stat -c %a {cookies}").strip() == "600"

    with subtest("whoami uses the stored server and session"):
        profile = json.loads(server.succeed("stashctl --json whoami"))
        assert profile["email"] == "alice@example.com", profile
        assert profile["name"] == "alice", profile
        assert "alice@example.com" in server.succeed("stashctl whoami")

    with subtest("an expired access token is refreshed"):
        stored = json.loads(server.succeed(f"cat {cookies}"))
        without_access = [
            cookie for cookie in stored
            if not cookie["raw_cookie"].startswith("stashsphere-access=")
        ]
        assert len(without_access) < len(stored), stored
        server.succeed(f"cat > {cookies} <<'EOF'\n{json.dumps(without_access)}\nEOF")
        profile = json.loads(server.succeed("stashctl --json whoami"))
        assert profile["email"] == "alice@example.com", profile
        refreshed = server.succeed(f"cat {cookies}")
        assert "stashsphere-access=" in refreshed, refreshed

    alice_id = stashctl_json("whoami")["id"]
    bob_id = next(user["id"] for user in api("GET", "/api/users") if user["name"] == "bob")

    drill = api("POST", "/api/things", {
        "name": "Cordless drill",
        "description": "18V with two batteries",
        "privateNote": "bought 2024",
        "imagesIds": [],
        "properties": [{"type": "float", "name": "weight", "value": 1.5, "unit": "kg"}],
        "quantity": 2,
        "quantityUnit": "pcs",
        "sharingState": "private",
    })
    workshop = api("POST", "/api/lists", {
        "name": "Workshop",
        "thingIds": [drill["id"]],
        "sharingState": "private",
    })
    share = api("POST", "/api/shares", {"targetUserId": bob_id, "objectId": drill["id"]})

    with subtest("things list and show"):
        things = stashctl_json("things list")
        assert things["totalCount"] == 1, things
        assert things["things"][0]["id"] == drill["id"], things
        assert stashctl_json("things list --search cord")["totalCount"] == 1
        assert stashctl_json("things list --search drill")["totalCount"] == 0
        assert stashctl_json("things list --all")["totalCount"] == 1

        table = server.succeed("stashctl things list")
        assert "Cordless drill" in table and "2 pcs" in table, table
        assert table.endswith("\n1 thing\n"), table

        thing = stashctl_json(f"things show {drill['id']}")
        assert thing["properties"] == [{"type": "float", "name": "weight", "value": 1.5, "unit": "kg"}], thing
        shown = server.succeed(f"stashctl things show {drill['id']}")
        for expected in ["Cordless drill", "Workshop", "bob", "weight:  1.5 kg", "bought 2024"]:
            assert expected in shown, f"{expected!r} missing in:\n{shown}"

    with subtest("missing objects exit with code 5"):
        status, out = server.execute("stashctl --json things show nope 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        assert json.loads(out)["error"]["status"] == 404, out

    with subtest("search finds things and lists"):
        result = stashctl_json("search batteries")
        assert [t["id"] for t in result["things"]] == [drill["id"]], result
        result = stashctl_json("search workshop")
        assert [l["id"] for l in result["lists"]] == [workshop["id"]], result
        assert "No things found." in server.succeed("stashctl search workshop")

    with subtest("lists list and show"):
        lists = stashctl_json("lists list")
        assert lists["totalCount"] == 1, lists
        shown = server.succeed(f"stashctl lists show {workshop['id']}")
        assert "Workshop" in shown and "Cordless drill" in shown, shown

    with subtest("users list and show"):
        users = stashctl_json("users list")
        assert {alice_id, bob_id} <= {u["id"] for u in users}, users
        assert stashctl_json(f"users show {bob_id}")["name"] == "bob"

    with subtest("shares show"):
        shown = stashctl_json(f"shares show {share['id']}")
        assert shown["type"] == "thing", shown
        assert shown["targetUser"]["id"] == bob_id, shown
        assert "shared with:  bob" in server.succeed(f"stashctl shares show {share['id']}")

    with subtest("notifications list for the receiving user"):
        # wait for the login rate limit to allow one more request
        time.sleep(12)
        server.succeed(
            f"echo secret | STASHCTL_HOME=/tmp/bob stashctl --server {server_url} "
            + "login --email bob@example.com --password-stdin"
        )
        notifications = stashctl_json("notifications list --unacknowledged", home="/tmp/bob")
        assert notifications["totalCount"] == 1, notifications
        notification = notifications["notifications"][0]
        assert notification["contentType"] == "THING_SHARED", notification
        assert notification["content"]["thingId"] == drill["id"], notification
        table = server.succeed("STASHCTL_HOME=/tmp/bob stashctl notifications list")
        assert f"thing {drill['id']} shared by user {alice_id}" in table, table

    with subtest("notifications acknowledge"):
        out = server.succeed(
            f"STASHCTL_HOME=/tmp/bob stashctl notifications ack {notification['id']}"
        )
        assert out.strip() == "Acknowledged 1 notification", out
        notifications = stashctl_json("notifications list --unacknowledged", home="/tmp/bob")
        assert notifications["totalCount"] == 0, notifications
        status, out = server.execute("STASHCTL_HOME=/tmp/bob stashctl --json notifications ack nope 2>&1")
        assert status != 0, f"exit code {status}: {out}"

    with subtest("things create"):
        saw = stashctl_json(
            "things create --name 'Hand saw' --quantity 1 --unit pcs "
            + "--property brand=Bahco --property 'length:float=50 cm' "
            + "--property sharp:boolean=true --property bought:datetime=2025-03-01"
        )
        thing = stashctl_json(f"things show {saw['id']}")
        assert thing["name"] == "Hand saw", thing
        assert thing["sharingState"] == "private", thing
        assert thing["properties"] == [
            {"type": "string", "name": "brand", "value": "Bahco"},
            {"type": "float", "name": "length", "value": 50, "unit": "cm"},
            {"type": "boolean", "name": "sharp", "value": True},
            {"type": "datetime", "name": "bought", "value": "2025-03-01T00:00:00Z"},
        ], thing
        out = server.succeed("stashctl things create --name 'Spare blade' --sharing friends")
        assert out.startswith("Created thing Spare blade ("), out
        status, out = server.execute("stashctl things create --name x --property w:float=abc 2>&1")
        assert status == 2, f"exit code {status}: {out}"

    with subtest("things update keeps the fields it does not change"):
        thing = stashctl_json(
            f"things update {drill['id']} --quantity 3 --property 'weight:float=1.7 kg' "
            + "--property color=green"
        )
        assert thing["quantity"] == 3, thing
        assert thing["name"] == "Cordless drill", thing
        assert thing["privateNote"] == "bought 2024", thing
        assert thing["quantityUnit"] == "pcs", thing
        assert thing["properties"] == [
            {"type": "float", "name": "weight", "value": 1.7, "unit": "kg"},
            {"type": "string", "name": "color", "value": "green"},
        ], thing
        thing = stashctl_json(f"things show {drill['id']}")
        assert [l["id"] for l in thing["lists"]] == [workshop["id"]], thing
        assert [s["id"] for s in thing["shares"]] == [share["id"]], thing

        shown = server.succeed(
            f"stashctl things update {saw['id']} --description 'pull saw' --remove-property brand"
        )
        assert "pull saw" in shown and "brand" not in shown, shown
        assert "length:  50 cm" in shown, shown

    with subtest("only the owner can change a thing"):
        status, out = server.execute(
            f"STASHCTL_HOME=/tmp/bob stashctl --json things update {drill['id']} --quantity 9 2>&1"
        )
        assert status == 4, f"exit code {status}: {out}"
        assert stashctl_json(f"things show {drill['id']}")["quantity"] == 3

    with subtest("lists create, update and delete"):
        garden = stashctl_json(f"lists create --name Garden --thing {saw['id']}")
        assert stashctl_json(f"lists show {garden['id']}")["sharingState"] == "private"
        out = server.succeed("stashctl lists create --name Empty")
        assert out.startswith("Created list Empty ("), out

        shed = stashctl_json(
            f"lists update {garden['id']} --name 'Garden shed' --sharing friends "
            + f"--add-thing {drill['id']} --remove-thing {saw['id']}"
        )
        assert shed["name"] == "Garden shed", shed
        assert shed["sharingState"] == "friends", shed
        assert [t["id"] for t in shed["things"]] == [drill["id"]], shed
        shed = stashctl_json(f"lists update {garden['id']} --add-thing {saw['id']}")
        assert shed["name"] == "Garden shed", shed
        assert sorted(t["id"] for t in shed["things"]) == sorted([drill["id"], saw["id"]]), shed

        out = server.succeed(f"stashctl lists delete {garden['id']}")
        assert out.strip() == f"Deleted list {garden['id']}", out
        status, out = server.execute(f"stashctl --json lists show {garden['id']} 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        stashctl_json(f"things show {saw['id']}")

    with subtest("shares create and delete"):
        list_share = stashctl_json(f"shares create {workshop['id']} --user {bob_id}")
        assert list_share["type"] == "list", list_share
        assert list_share["targetUser"]["id"] == bob_id, list_share
        shown = stashctl_json(f"lists show {workshop['id']}", home="/tmp/bob")
        assert shown["name"] == "Workshop", shown

        server.succeed(f"stashctl shares delete {list_share['id']}")
        status, out = server.execute(f"STASHCTL_HOME=/tmp/bob stashctl --json lists show {workshop['id']} 2>&1")
        assert status == 4, f"exit code {status}: {out}"
        status, out = server.execute(f"stashctl --json shares show {list_share['id']} 2>&1")
        assert status == 5, f"exit code {status}: {out}"

    with subtest("things delete"):
        out = server.succeed(f"stashctl things delete {saw['id']}")
        assert out.strip() == f"Deleted thing {saw['id']}", out
        status, out = server.execute(f"stashctl --json things show {saw['id']} 2>&1")
        assert status == 5, f"exit code {status}: {out}"

    with subtest("STASHCTL_OUTPUT selects JSON, compact when not on a terminal"):
        out = server.succeed("STASHCTL_OUTPUT=json stashctl whoami")
        assert out.count("\n") == 1, out
        assert json.loads(out)["email"] == "alice@example.com", out

    with subtest("logout removes the session"):
        server.succeed("stashctl logout")
        server.fail(f"test -e {cookies}")
        status, out = server.execute("stashctl --json whoami 2>&1")
        assert status == 3, f"exit code {status}: {out}"
  '';
}
