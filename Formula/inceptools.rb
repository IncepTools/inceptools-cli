class Inceptools < Formula
  desc "A powerful, developer-friendly database migration CLI for Go"
  homepage "https://github.com/IncepTools/inceptools-cli"
  url "https://github.com/IncepTools/inceptools-cli/releases/download/v{{VERSION}}/inceptools-{{OS}}-{{ARCH}}.tar.gz"
  sha256 "{{SHA256}}"
  license "GPL-3.0"

  def install
    bin.install "inceptools"
  end

  test do
    system "#{bin}/inceptools", "version"
  end
end
