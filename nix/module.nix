{ config, lib, pkgs, ... }:
let
  cfg = config.services.nwxraytui;
in {
  options.services.nwxraytui = {
    enable = lib.mkEnableOption "nwxraytui proxy daemon";

    enableTun = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = ''
        Grant CAP_NET_ADMIN to the daemon so TUN mode can be enabled from
        within the TUI. Follows the same pattern as services.mihomo.enableTun.
      '';
    };

    user = lib.mkOption {
      type = lib.types.str;
      description = "User account to run the nwxraytui daemon as.";
    };

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.nwxraytui;
      description = "The nwxraytui package to use.";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.nwxraytui = {
      description = "nwxraytui proxy daemon";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];

      serviceConfig = {
        # Create /run/user/%U/nwxraytui as root before dropping to User=;
        # required because /run/user/{uid}/ may not exist at boot.
        ExecStartPre = "+${pkgs.coreutils}/bin/install -d -m 0700 -o %u -g %g /run/user/%U /run/user/%U/nwxraytui";
        ExecStart = "${cfg.package}/bin/nwxraytui --daemon";
        Restart = "on-failure";
        RestartSec = "5s";
        User = cfg.user;
      } // lib.optionalAttrs cfg.enableTun {
        AmbientCapabilities = [ "CAP_NET_ADMIN" ];
        CapabilityBoundingSet = [ "CAP_NET_ADMIN" ];
      };
    };
  };
}
