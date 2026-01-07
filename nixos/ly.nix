{ pkgs, inputs, config, ... }:
{
  # Enable Ly (TTY-based display manager).
  # Ly is lightweight and works well with Wayland/X sessions provided
  # by installed desktop packages (e.g. Hyprland). This module keeps
  # the config minimal — Ly will use the system's session files.
  services.displayManager = {
    ly = {
      enable = true;
      package = pkgs.ly;
      # Keep Ly's own autologin off so users must enter their password.
      # Ly will present the session/user chooser by default.
      # Use the system-wide default session to point users to Hyprland.
      # `extraConfig` can still be used for finer theming later.
      # extraConfig = ''
      #   # example ly config options
      # '';
    };
    # Do not enable automatic autologin here; keep manual login.
    autoLogin = {
      enable = false;
    };
    # Prefer Hyprland as the default session shown in the chooser.
    defaultSession = "hyprland";
  };

  # Ensure the ly package is available on the system path
  environment.systemPackages = with pkgs; [ ly ];
}
