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

    server_url = "http://127.0.0.1:8081"
    cookies = "/tmp/stashctl/cookies.json"

    start_all()
    server.wait_for_unit("stashsphere.service")
    server.wait_until_succeeds(f"${pkgs.curl}/bin/curl -f {server_url}/api/info")

    alice = json.dumps({
        "name": "alice",
        "email": "alice@example.com",
        "password": "secret",
        "inviteCode": "1234",
    })
    server.succeed(
        "${pkgs.curl}/bin/curl -f -H 'content-type: application/json' "
        + f"-d '{alice}' {server_url}/api/user/register"
    )

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

    with subtest("logout removes the session"):
        server.succeed("stashctl logout")
        server.fail(f"test -e {cookies}")
        status, out = server.execute("stashctl --json whoami 2>&1")
        assert status == 3, f"exit code {status}: {out}"
  '';
}
