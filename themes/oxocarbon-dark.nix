{
  lib,
  pkgs,
  config,
  ...}: {
  options.theme = lib.mkOption {
    type = lib.types.attrs;
    default = {
      rounding = 5;
      gaps-in = 12;
      gaps-out = 12 * 2;
      active-opacity = 0.80;
      inactive-opacity = 0.50;
      blur = true;
      border-size = 1;
      animation-speed = "fast"; # "fast" | "medium" | "slow"
      fetch = "none"; # "nerdfetch" | "neofetch" | "pfetch" | "none"
      textColorOnWallpaper = config.lib.stylix.colors.base00;
    };
    description = "Theme configuration options";
  };

  config.stylix = {
    enable = true;

    base16Scheme = {
      base00 = "161616";
      base01 = "262626";
      base02 = "393939";
      base03 = "525252";
      base04 = "dde1e6";
      base05 = "f2f4f8";
      base06 = "ffffff";
      base07 = "08bdba";
      base08 = "ee5396";
      base09 = "ff7eb6";
      base0A = "ff6f00";
      base0B = "42be65";
      base0C = "3ddbd9";
      base0D = "33b1ff";
      base0E = "be95ff";
      base0F = "82cfff";
    };

    cursor = {
      name = "BreezeX-RosePine-Linux";
      package = pkgs.rose-pine-cursor;
      size = 30;
    };

    fonts = {
      monospace = {
        package = pkgs.nerd-fonts.jetbrains-mono;
        name = "JetBrains Mono Nerd Font";
      };
      sansSerif = {
        package = pkgs.source-sans-pro;
        name = "Source Sans Pro";
      };
      serif = config.stylix.fonts.sansSerif;
      emoji = {
        package = pkgs.noto-fonts-color-emoji;
        name = "Noto Color Emoji";
      };
      sizes = {
        applications = 14;
        desktop = 14;
        popups = 14;
        terminal = 14;
      };
    };

    polarity = "dark";

    image = pkgs.fetchurl {
    url = "https://raw.githubusercontent.com/anotherhadi/awesome-wallpapers/refs/heads/main/app/static/wallpapers/tokyo-skyscraper_dark.png";
    sha256 = "1kg2rglfhsm725y2ni634fx9a837qn23s5ba4jxdwxfjc56hgla9";
    };
  };
}
