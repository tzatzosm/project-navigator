class ProjectNavigator < Formula
  desc "CLI for managing and navigating development projects"
  homepage "https://github.com/tzatzosm/project-navigator"
  url "https://github.com/tzatzosm/project-navigator/archive/refs/tags/v0.2.0.tar.gz"
  # TODO: after pushing the v0.2.0 tag (the Go rewrite), fill this in with:
  #   curl -sL <url above> | shasum -a 256
  sha256 "REPLACE_WITH_RELEASE_TARBALL_SHA256"
  license "MIT"
  head "https://github.com/tzatzosm/project-navigator.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(output: bin/"pn", ldflags: "-s -w"), "./cmd/pn"
  end

  test do
    assert_match "usage: pn", shell_output("#{bin}/pn --help")
    output = shell_output("HOME=#{testpath} #{bin}/pn list 2>&1")
    assert_match "No projects", output
  end
end
