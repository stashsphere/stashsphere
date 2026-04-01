# Development environment using process-compose.
# Orchestrates PostgreSQL and Dex as Nix-built services,
# while backend and frontend run from source for fast iteration.
# Usage: nix run .#dev
{ pkgs }:
let
  pgPort = "5433";
  dexPort = "5557";
  backendPort = "8082";
  frontendPort = "5173";
  frontendNixPort = "5174";

  dexConfig = import ./dex-config.nix;

  dexSettingsFile = pkgs.writeText "dex-dev-config.yaml" (builtins.toJSON
    (dexConfig.dexSettings {
      issuerBase = "http://localhost:${dexPort}";
      dexListenAddr = "localhost:${dexPort}";
      backendBases = [
        "http://localhost:${backendPort}"
      ];
    })
  );

  backendConfigFile = pkgs.writeText "backend-dev.json" (builtins.toJSON {
    listenAddress = ":${backendPort}";
    baseUrl = "http://localhost:${backendPort}";
    database = {
      host = "127.0.0.1";
      port = builtins.fromJSON pgPort;
      user = "stashsphere";
      password = "secret";
      dbname = "stashsphere";
      sslmode = "disable";
    };
    invites = {
      enabled = true;
      code = "1234";
    };
    domains = {
      allowed = [ "http://localhost:${frontendPort}" "http://localhost:${frontendNixPort}" ];
      cookieDomain = "localhost";
    };
    email = {
      backend = "stdout";
    };
    auth = {
      disableSecureCookies = true;
      oidc = {
        enabled = true;
        providers = [
          {
            name = "dex";
            display_name = "Dex";
            issuer_url = "http://localhost:${dexPort}/dex";
            client_id = dexConfig.oidcClientConfig.clientId;
            client_secret = dexConfig.oidcClientConfig.clientSecret;
            scopes = [ "openid" "profile" "email" ];
          }
        ];
      };
    };
  });

  postgresConf = pkgs.writeText "postgresql.conf" ''
    listen_addresses = '127.0.0.1'
    port = ${pgPort}
  '';

  processComposeConfig = pkgs.writeText "process-compose.yaml" (builtins.toJSON {
    version = "0.5";

    processes = {
      postgres = {
        command = pkgs.writeShellScript "postgres-start" ''
          set -euo pipefail
          export PGDATA="$DEV_DATA_DIR/postgres"
          if [ ! -d "$PGDATA" ]; then
            ${pkgs.postgresql}/bin/initdb -D "$PGDATA" --auth=scram-sha-256 -U stashsphere --pwfile=${pkgs.writeText "pgpass" "secret"}
          fi
          install -m 644 ${postgresConf} "$PGDATA/postgresql.conf"
          exec ${pkgs.postgresql}/bin/postgres -D "$PGDATA" -k "$DEV_DATA_DIR"
        '';
        readiness_probe = {
          exec.command = "${pkgs.postgresql}/bin/pg_isready -h 127.0.0.1 -p ${pgPort} -U stashsphere -d postgres";
          initial_delay_seconds = 1;
          period_seconds = 2;
        };
        shutdown.signal = 15;
        availability.restart = "on_failure";
      };

      postgres-init = {
        command = pkgs.writeShellScript "postgres-init" ''
          set -euo pipefail
          export PGPASSWORD=secret
          ${pkgs.postgresql}/bin/psql -h 127.0.0.1 -p ${pgPort} -U stashsphere -d postgres \
            -tc "SELECT 1 FROM pg_database WHERE datname='stashsphere'" | grep -q 1 || \
            ${pkgs.postgresql}/bin/createdb -h 127.0.0.1 -p ${pgPort} -U stashsphere stashsphere
        '';
        depends_on.postgres.condition = "process_healthy";
        availability.restart = "no";
      };

      dex = {
        command = pkgs.writeShellScript "dex-start" ''
          set -euo pipefail
          mkdir -p "$DEV_DATA_DIR/dex"
          DEX_CONFIG="$DEV_DATA_DIR/dex/config.yaml"
          ${pkgs.jq}/bin/jq --arg db "$DEV_DATA_DIR/dex/dex.db" \
            '.storage = {"type": "sqlite3", "config": {"file": $db}}' \
            ${dexSettingsFile} > "$DEX_CONFIG"
          exec ${pkgs.dex-oidc}/bin/dex serve "$DEX_CONFIG"
        '';
        readiness_probe = {
          exec.command = "${pkgs.curl}/bin/curl -sf http://127.0.0.1:${dexPort}/dex/.well-known/openid-configuration";
          initial_delay_seconds = 2;
          period_seconds = 3;
        };
        availability.restart = "on_failure";
      };

      backend = {
        command = pkgs.writeShellScript "backend-start" ''
          set -euo pipefail

          AUTH_KEY_FILE="$DEV_DATA_DIR/auth.key"

          if [ ! -f "$AUTH_KEY_FILE" ]; then
            pushd backend
            ${pkgs.go}/bin/go run ./... genkey > "$AUTH_KEY_FILE"
            echo "Generated new auth secret at $AUTH_KEY_FILE"
            popd
          fi

          export STASHSPHERE_AUTH__PRIVATE_KEY="$(cat "$AUTH_KEY_FILE")"

          cd backend
          ${pkgs.go}/bin/go run ./... migrate --conf ${backendConfigFile}
          exec ${pkgs.go}/bin/go run ./... serve --conf ${backendConfigFile} --serve-openapi
        '';
        depends_on = {
          postgres-init.condition = "process_completed_successfully";
        };
        readiness_probe = {
          exec.command = "${pkgs.curl}/bin/curl -sf http://127.0.0.1:${backendPort}/api/info";
          initial_delay_seconds = 3;
          period_seconds = 3;
        };
        availability.restart = "on_failure";
        environment = [
          "CGO_ENABLED=1"
          "CGO_CFLAGS=-I${pkgs.file.dev}/include"
          "CGO_LDFLAGS=-L${pkgs.file}/lib"
        ];
      };

      frontend = {
        command = pkgs.writeShellScript "frontend-start" ''
          set -euo pipefail
          cd frontend
          if [ ! -d node_modules ]; then
            ${pkgs.pnpm}/bin/pnpm install
          fi
          exec ${pkgs.pnpm}/bin/pnpm dev -- --port ${frontendPort}
        '';
        depends_on = {
          backend.condition = "process_healthy";
        };
        availability.restart = "on_failure";
      };

      frontend-nix = {
        command =
          let
            frontend-dev = pkgs.stashsphere-frontend.override {
              apiHost = "http://localhost:${backendPort}";
            };
          in
          pkgs.writeShellScript "frontend-nix-start" ''
            set -euo pipefail
            echo "Serving Nix-built frontend on port ${frontendNixPort}..."
            cd ${frontend-dev}/dist
            exec ${pkgs.nodePackages.serve}/bin/serve -s -l ${frontendNixPort}
          '';
        depends_on = {
          backend.condition = "process_healthy";
        };
        availability.restart = "on_failure";
      };
    };
  });
in
pkgs.writeShellApplication {
  name = "stashsphere-dev";
  runtimeInputs = [
    pkgs.process-compose
    pkgs.file
  ];
  text = ''
    export DEV_DATA_DIR="''${DEV_DATA_DIR:-$PWD/.dev-data}"
    mkdir -p "$DEV_DATA_DIR"
    echo "Development data directory: $DEV_DATA_DIR"
    echo "PostgreSQL: 127.0.0.1:${pgPort}"
    echo "Dex OIDC:   http://localhost:${dexPort}/dex"
    echo "Backend:    http://localhost:${backendPort}"
    echo "Frontend:   http://localhost:${frontendPort} (Vite default)"
    echo "Frontend (Nix): http://localhost:${frontendNixPort}"
    exec process-compose up -f ${processComposeConfig} --port 9847
  '';
}
