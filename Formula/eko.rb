class Eko < Formula
  desc "Eko: AI-powered snapshot versioning CLI"
  homepage "https://github.com/kavix/eko"
  url "https://github.com/kavix/eko/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256" # This will be updated by GoReleaser
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"eko", "main.go"
  end

  test do
    system "#{bin}/eko", "--help"
  end
end
