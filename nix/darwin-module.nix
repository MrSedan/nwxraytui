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
        Run the daemon as root so TUN (utun) mode can be enabled from the TUI.
        On macOS, TUN requires root privileges.
      '';
    };

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.nwxraytui;
      description = "The nwxraytui package to use.";
    };
  };

  config = lib.mkIf cfg.enable {
    launchd.user.agents.nwxraytui = lib.mkIf (!cfg.enableTun) {
      serviceConfig = {
        Label = "dev.nwxraytui.daemon";
        ProgramArguments = [ "${cfg.package}/bin/nwxraytui" "--daemon" ];
        RunAtLoad = true;
        KeepAlive = true;
      };
    };

    launchd.daemons.nwxraytui = lib.mkIf cfg.enableTun {
      serviceConfig = {
        Label = "dev.nwxraytui.daemon";
        ProgramArguments = [ "${cfg.package}/bin/nwxraytui" "--daemon" ];
        RunAtLoad = true;
        KeepAlive = true;
        UserName = "root";
      };
    };
  };
}
