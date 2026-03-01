# Shared Dex OIDC configuration for development and testing.
# Used by both the NixOS VM test and the process-compose dev environment.
{
  # Dex settings that map to its YAML configuration.
  dexSettings =
    { issuerBase ? "http://127.0.0.1:5556"
    , dexListenAddr ? "0.0.0.0:5556"
    , backendBases ? [ "http://127.0.0.1:8081" "http://localhost:8081" ]
    , storagePath ? "/var/lib/dex/dex.db"
    }:
    {
      issuer = "${issuerBase}/dex";

      storage = {
        type = "sqlite3";
        config.file = storagePath;
      };

      web.http = dexListenAddr;

      staticClients = [
        {
          id = "stashsphere";
          name = "StashSphere";
          secret = "stashsphere-test-secret";
          redirectURIs = map (base: "${base}/api/auth/oidc/dex/callback") backendBases;
        }
      ];

      enablePasswordDB = true;

      staticPasswords = [
        {
          email = "alice@example.com";
          # bcrypt hash of "password"
          hash = "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W";
          username = "alice";
          userID = "08a8684b-db88-4b73-90a9-3cd1661f5466";
        }
        {
          email = "bob@example.com";
          # bcrypt hash of "password"
          hash = "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W";
          username = "bob";
          userID = "41331323-6f44-45e6-b3b9-2c4b60ce40a5";
        }
        {
          email = "charlie@example.com";
          # bcrypt hash of "password"
          hash = "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W";
          username = "charlie";
          userID = "99d7e095-742e-4dc3-b77f-b01df1e78c37";
        }
      ];
    };

  # Client credentials for StashSphere to connect to Dex
  oidcClientConfig = {
    clientId = "stashsphere";
    clientSecret = "stashsphere-test-secret";
  };
}
