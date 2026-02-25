{ testers, pkgs }:

let
  dexConfig = import ../../nix/dex-config.nix;
  dexSettings = dexConfig.dexSettings { issuerBase = "http://127.0.0.1:5556"; };
in
testers.nixosTest {
  name = "stashsphere-oidc";

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

    services.dex = {
      enable = true;
      settings = dexSettings;
    };

    systemd.services.dex.serviceConfig.StateDirectory = "dex";
  };

  testScript = ''
    start_all()

    server.wait_for_unit("dex.service")
    server.wait_for_unit("stashsphere.service")

    # Verify StashSphere is responding
    server.wait_until_succeeds("${pkgs.curl}/bin/curl http://127.0.0.1:8081")

    # Verify Dex OIDC discovery endpoint is responding
    server.wait_until_succeeds(
        "${pkgs.curl}/bin/curl -f http://127.0.0.1:5556/dex/.well-known/openid-configuration"
    )
  '';
}
