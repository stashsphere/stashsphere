# Shared Dex OIDC configuration for development and testing.
# Used by both the NixOS VM test and the process-compose dev environment.
{
  # Dex settings that map to its YAML configuration.
  # The issuerBase should be overridden per-environment if needed.
  dexSettings =
    { issuerBase ? "http://127.0.0.1:5556"
    , storagePath ? "/var/lib/dex/dex.db"
    }:
    {
      issuer = "${issuerBase}/dex";

      storage = {
        type = "sqlite3";
        config.file = storagePath;
      };

      web.http = "0.0.0.0:5556";

      staticClients = [
        {
          id = "stashsphere";
          name = "StashSphere";
          secret = "stashsphere-test-secret";
          redirectURIs = [
            "http://127.0.0.1:8081/auth/oidc/callback"
            "http://localhost:8081/auth/oidc/callback"
          ];
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
      ];
    };

  # Client credentials for StashSphere to connect to Dex
  oidcClientConfig = {
    clientId = "stashsphere";
    clientSecret = "stashsphere-test-secret";
  };
}
