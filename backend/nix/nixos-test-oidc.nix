{ testers, pkgs }:

let
  dexConfig = import ../../nix/dex-config.nix;
  dexSettings = dexConfig.dexSettings { issuerBase = "http://127.0.0.1:5556"; };
  oidcClientConfig = dexConfig.oidcClientConfig;

  testPy = ./oidc-test.py;

  python = pkgs.python3.withPackages (ps: [ ps.requests ps.beautifulsoup4 ]);
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
          auth = {
            disableSecureCookies = true;
            oidc = {
              enabled = true;
              providers = [
                {
                  name = "dex";
                  display_name = "Dex";
                  issuer_url = "http://127.0.0.1:5556/dex";
                  client_id = oidcClientConfig.clientId;
                  client_secret = oidcClientConfig.clientSecret;
                }
              ];
            };
          };
          domains = {
            allowed = [ "http://localhost:3000" ];
          };
          baseUrl = "http://127.0.0.1:8081";
          frontendUrl = "http://localhost:3000";
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
      settings = dexSettings // {
        oauth2 = { skipApprovalScreen = true; };
      };
    };

    systemd.services.dex.serviceConfig.StateDirectory = "dex";
  };

  testScript = ''
    start_all()

    server.wait_for_unit("dex.service")
    server.wait_for_unit("stashsphere.service")

    # Verify both services are responding
    server.wait_until_succeeds("${pkgs.curl}/bin/curl -f http://127.0.0.1:8081/api/info")
    server.wait_until_succeeds(
        "${pkgs.curl}/bin/curl -f http://127.0.0.1:5556/dex/.well-known/openid-configuration"
    )

    # Run the OIDC test suite
    server.succeed("${python}/bin/python3 ${testPy}")
  '';
}
