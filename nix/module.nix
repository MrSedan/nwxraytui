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

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.nwxraytui;
      description = "The nwxraytui package to use.";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.user.services.nwxraytui = {
      description = "nwxraytui proxy daemon";
      after = [ "network.target" ];
      wantedBy = [ "default.target" ];

      serviceConfig = {
        ExecStart = "${cfg.package}/bin/nwxraytui --daemon";
        Restart = "on-failure";
        RestartSec = "5s";
      } // lib.optionalAttrs cfg.enableTun {
        AmbientCapabilities = [ "CAP_NET_ADMIN" ];
        CapabilityBoundingSet = [ "CAP_NET_ADMIN" ];
      };
    };
  };
}
