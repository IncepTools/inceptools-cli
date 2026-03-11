class Inceptools < Formula
  desc "A powerful, developer-friendly database migration CLI for Go"
  homepage "https://github.com/IncepTools/inceptools-cli"
  version "{{VERSION}}"
  license "GPL-3.0"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/IncepTools/inceptools-cli/releases/download/v#{version}/inceptools-darwin-amd64.tar.gz"
      sha256 "{{SHA256_DARWIN_AMD64}}"
    else
      url "https://github.com/IncepTools/inceptools-cli/releases/download/v#{version}/inceptools-darwin-arm64.tar.gz"
      sha256 "{{SHA256_DARWIN_ARM64}}"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/IncepTools/inceptools-cli/releases/download/v#{version}/inceptools-linux-amd64.tar.gz"
      sha256 "{{SHA256_LINUX_AMD64}}"
    else
      url "https://github.com/IncepTools/inceptools-cli/releases/download/v#{version}/inceptools-linux-arm64.tar.gz"
      sha256 "{{SHA256_LINUX_ARM64}}"
    end
  end

  def install
    # Rename the platform-specific binary to 'inceptools'
    bin.install Dir["inceptools-*"][0] => "inceptools"
  end

  test do
    system "#{bin}/inceptools", "version"
  end
end
