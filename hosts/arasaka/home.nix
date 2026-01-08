{
  pkgs,
  config,
  ...
}: {
  imports = [
    # Programs
    ../../home/programs/nvf
    ../../home/programs/shell
    ../../home/programs/fetch
    ../../home/programs/git
    ../../home/programs/git/lazygit.nix
    ../../home/programs/thunar
    ../../home/programs/discord
    ../../home/programs/nixy
    ../../home/programs/zathura


    # System (Desktop environment like stuff)
    ../../home/system/hyprland
    ../../home/system/hyprpaper
    ../../home/system/mime
    ../../home/system/udiskie

    ./variables.nix # Mostly user-specific configuration
    #./secrets # CHANGEME: You should probably remove this line, this is where I store my secrets
  ];

  home = {
    packages = with pkgs; [
      # Apps
      vlc # Video player
      blanket # White-noise app
      obsidian # Note taking app
      textpieces # Manipulate texts
      resources # Ressource monitor
      gnome-clocks # Clocks app
      gnome-text-editor # Basic graphic text editor
      mpv # Video player
      signal-desktop # Signal app, private messages
      solaar # Logitech devices manager


      # Dev
      go
      bun
      docker
      nodejs
      python3
      jq
      just
      pnpm
      air
      duckdb

      # Just cool
      peaclock
      cbonsai
      pipes
      cmatrix
      fastfetch

      # Backup
      vscode
    ];

    inherit (config.var) username;
    homeDirectory = "/home/" + config.var.username;


    # Home Manager compatibility pin — controls module defaults/migrations
    stateVersion = "24.05";
  };

  programs.home-manager.enable = true;
}
