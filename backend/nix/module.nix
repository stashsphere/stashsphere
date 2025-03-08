{ config
, lib
, pkgs
, ...
}:
let
  cfg = config.services.stashsphere;
in
{
  options.services.stashsphere = {
    enable = lib.mkEnableOption "StashSphere Inventory Service";
    listenAddress = lib.mkOption {
      type = lib.types.str;
      default = ":8081";
      description = "Address and port to expose api";
    };
    settings = lib.mkOption {
      type = lib.types.attrs;
      default = { };
      description = "Settings for StashSphere";
    };
    configFiles = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      description = "List of files to include, use for secrets";
    };
    usesLocalPostgresql = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = "Whether stashsphere will connect to a local postgresql server.";
    };
  };

  config = lib.mkIf cfg.enable
    {
      systemd.services.stashsphere =
        let
          settingsFile = pkgs.writeText "settings.json" (builtins.toJSON cfg.settings);
          configFilesArgs = builtins.concatStringsSep " --conf " (cfg.configFiles ++ [ settingsFile ]);
        in
        {
          wantedBy = [ "multi-user.target" ];
          after = [ "network.target" ] ++ (if cfg.usesLocalPostgresql then [ "postgresql.service" ] else [ ]);
          serviceConfig = {
            Restart = "always";
            DynamicUser = true;
            MemoryDenyWriteExecute = true;
            PrivateDevices = true;
            ProtectSystem = "strict";
            RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX"];
            RestrictNamespaces = true;
            RestrictSUIDSGID = true;
            ExecStartPre = ''
              ${pkgs.stashsphere}/bin/backend migrate --conf ${settingsFile} ${configFilesArgs}
            '';
            ExecStart = ''
              ${pkgs.stashsphere}/bin/backend serve --conf ${configFilesArgs}
            '';
            User="stashsphere";
            StateDirectory="stashsphere";
            CacheDirectory="stashsphere";
            RuntimeDirectory="stashsphere";
          };
        };
    };
}
