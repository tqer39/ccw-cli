class Ccw < Formula
  desc "Launch Claude Code in an isolated git worktree"
  homepage "https://github.com/tqer39/ccw-cli"
  url "https://github.com/tqer39/ccw-cli/archive/refs/tags/v0.20.0.tar.gz"
  sha256 "52b315ed2fffc1e4c15fe68851b67475c1149650ba18df03a0b22dd82cc6e5a7"
  license "MIT"
  head "https://github.com/tqer39/ccw-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X github.com/tqer39/ccw-cli/internal/version.Version=#{version}
      -X github.com/tqer39/ccw-cli/internal/version.Commit=brew
      -X github.com/tqer39/ccw-cli/internal/version.Date=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/ccw"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/ccw -v")
    system bin/"ccw", "-h"
  end
end
