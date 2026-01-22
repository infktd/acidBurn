# nixpkgs Submission Checklist

## Before You Start

- [ ] Push v0.1.0 tag to GitHub
- [ ] Create GitHub release with changelog
- [ ] Verify `nix run github:infktd/devdash` works
- [ ] Test on both Linux and macOS if possible

## Fork and Setup

```bash
# 1. Fork nixpkgs on GitHub
# https://github.com/NixOS/nixpkgs -> Fork

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/nixpkgs
cd nixpkgs

# 3. Add upstream remote
git remote add upstream https://github.com/NixOS/nixpkgs

# 4. Create branch
git checkout -b devdash
```

## Create Package File

```bash
# Create directory
mkdir -p pkgs/by-name/de/devdash

# Create package.nix (see below)
```

### Package File: `pkgs/by-name/de/devdash/package.nix`

```nix
{ lib
, buildGoModule
, fetchFromGitHub
}:

buildGoModule rec {
  pname = "devdash";
  version = "0.1.0";

  src = fetchFromGitHub {
    owner = "infktd";
    repo = "devdash";
    rev = "v${version}";
    hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Compute with nix-prefetch
  };

  vendorHash = "sha256-CwWLZBudC+cUEDWZAqhOhjwLYybjATrC9VG7GNVLNnQ=";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Terminal dashboard for managing devenv.sh projects";
    homepage = "https://github.com/infktd/devdash";
    changelog = "https://github.com/infktd/devdash/releases/tag/v${version}";
    license = licenses.mit;
    maintainers = with maintainers; [ ]; # Add your GitHub handle here
    mainProgram = "devdash";
    platforms = platforms.unix;
  };
}
```

## Compute Source Hash

```bash
# From nixpkgs directory
nix-prefetch-url --unpack https://github.com/infktd/devdash/archive/v0.1.0.tar.gz

# Or use nix-prefetch-github (if installed)
nix-prefetch-github infktd devdash --rev v0.1.0

# Copy the hash and update package.nix
```

## Test the Package

```bash
# Build
nix-build -A devdash

# Should create ./result/bin/devdash

# Test binary
./result/bin/devdash --help

# Check metadata
nix-instantiate --eval -E 'with import ./. {}; devdash.meta.description'

# Format code
nix-shell -p nixpkgs-fmt --run "nixpkgs-fmt pkgs/by-name/de/devdash/package.nix"

# Run package tests (if any)
nix-build -A devdash.tests
```

## Add Yourself as Maintainer (Optional)

Edit `maintainers/maintainer-list.nix`:

```nix
  your-github-handle = {
    email = "your@email.com";
    github = "your-github-handle";
    githubId = 12345678; # Your GitHub user ID
    name = "Your Name";
  };
```

Then update package.nix:
```nix
maintainers = with maintainers; [ your-github-handle ];
```

## Commit and Push

```bash
# Stage changes
git add pkgs/by-name/de/devdash/package.nix
git add maintainers/maintainer-list.nix  # If you added yourself

# Commit with proper message
git commit -m "devdash: init at 0.1.0"

# Push to your fork
git push origin devdash
```

## Create Pull Request

1. Go to: https://github.com/NixOS/nixpkgs/compare
2. Select: `base: master` ← `compare: YOUR_USERNAME:devdash`
3. Title: `devdash: init at 0.1.0`
4. Copy description from `docs/NIXPKGS_PR.md`
5. Create PR

## After PR is Created

- [ ] Check CI/CD results (OfBorg will test builds)
- [ ] Respond to reviewer comments within 24-48 hours
- [ ] Make requested changes promptly
- [ ] Be patient - review can take 1-2 weeks

## OfBorg Commands

Reviewers will use these commands to test your package:

```
@ofborg build devdash         # Build on all platforms
@ofborg test devdash           # Run package tests
```

You might see comments like:
- "Built successfully on x86_64-linux"
- "Built successfully on aarch64-darwin"
- "No tests to run"

## Common Review Feedback

Be prepared to address:
- **License verification**: Ensure LICENSE file matches `meta.license`
- **Description clarity**: Make sure it's understandable to non-devenv users
- **Platform tags**: Verify `platforms.unix` is appropriate
- **Maintainer info**: Double-check email and GitHub ID
- **Code formatting**: Run `nixpkgs-fmt` before committing
- **Unnecessary dependencies**: Remove any that aren't needed

## Helpful Resources

- [Nixpkgs Manual - Quick Start](https://nixos.org/manual/nixpkgs/stable/#chap-quick-start)
- [Contributing Guidelines](https://github.com/NixOS/nixpkgs/blob/master/CONTRIBUTING.md)
- [Go Packages in nixpkgs](https://nixos.org/manual/nixpkgs/stable/#sec-language-go)
- [nixpkgs Pull Request Template](https://github.com/NixOS/nixpkgs/blob/master/.github/PULL_REQUEST_TEMPLATE.md)

## Timeline Expectations

- **PR Creation**: 5 minutes
- **OfBorg Initial Build**: 10-30 minutes
- **First Reviewer Comment**: 1-7 days
- **Back-and-forth Reviews**: 3-14 days
- **Merge**: 1-4 weeks total (if no major issues)

## After Merge

- [ ] Update README to mention nixpkgs availability
- [ ] Announce on Discourse/Reddit
- [ ] Thank reviewers and maintainers
- [ ] Consider maintaining the package (respond to updates/issues)

Good luck! 🚀
