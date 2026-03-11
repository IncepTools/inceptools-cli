class Inceptools < Formula
  desc "A powerful, developer-friendly database migration CLI for Go"
  homepage "https://github.com/IncepTools/inceptools-cli"
  version "1.0.1"
  license "GPL-3.0"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/IncepTools/inceptools-cli/releases/download/v#{version}/inceptools-darwin-amd64.tar.gz"
      sha256 "c46420b67b86222b52d1565e7a1aa525eca32d5a605f658b9c766153b584c015"
    else
      url "https://github.com/IncepTools/inceptools-cli/releases/download/v#{version}/inceptools-darwin-arm64.tar.gz"
      sha256 "8e93108732740be75735d5ce6f750408576ff2242bc44478ed6dc592d2f6d467"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/IncepTools/inceptools-cli/releases/download/v#{version}/inceptools-linux-amd64.tar.gz"
      sha256 "8e77a24163a09c7833600a5902b268ab6e53d5c5112a2c4fd6950143494ad074"
    else
      url "https://github.com/IncepTools/inceptools-cli/releases/download/v#{version}/inceptools-linux-arm64.tar.gz"
      sha256 "569febffeacea2448b65b319fabf717f8873f152eac31a07810ab4fe4b3ce549"
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
