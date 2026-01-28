class Dotcor < Formula
  desc "Symlink-based dotfile manager with Git integration"
  homepage "https://github.com/justincordova/dotcor"
  url "https://github.com/justincordova/dotcor/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "c92241f4dd4003b40cc631a48ee0cd166a7171fb6d9fc6e8e1a4ed877e784d04"
  license "MIT"
  head "https://github.com/justincordova/dotcor.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/dotcor"
  end

  test do
    assert_match "dotcor version #{version}", shell_output("#{bin}/dotcor --version")
  end
end
