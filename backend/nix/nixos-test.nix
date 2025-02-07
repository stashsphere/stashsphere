{ nixosTest, pkgs }:

nixosTest {
  name = "stashsphere";

  nodes.server = { ... }: {
    imports = [ ./module.nix ];
    services.stashsphere =
      let
        secretConfig = pkgs.writeText "secret.json" (builtins.toJSON {
          invites = {
            enabled = true;
            code = "1234";
          };
        });
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
  };

  testScript = ''
    start_all()
    server.wait_for_unit("stashsphere.service")
    server.wait_until_succeeds("${pkgs.curl}/bin/curl http://127.0.0.1:8081")
  '';
}
