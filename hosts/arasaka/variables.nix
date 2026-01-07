{
  config,
  lib,
  ...
}: {
  imports = [
    # Choose your theme here:
    ../../themes/rose-pine.nix
  ];

  config.var = {
    hostname = "arasaka";
    username = "infktd";
    configDirectory =
      "/home/"
      + config.var.username
      + "/.config/nixos"; # The path of the nixos configuration directory

  keyboardLayout = "us";

    location = "Dallas";
  timeZone = "America/Chicago";
  defaultLocale = "en_US.UTF-8";
  # Additional locale used for various LC_* settings (leave same as default if not needed)
  extraLocale = "en_US.UTF-8";

    git = {
      username = "infktd";
      email = "112569860+anotherhadi@users.noreply.github.com";
    };

    autoUpgrade = false;
    autoGarbageCollector = true;
  };

  # DON'T TOUCH THIS
  options = {
    var = lib.mkOption {
      type = lib.types.attrs;
      default = {};
    };
  };
}
