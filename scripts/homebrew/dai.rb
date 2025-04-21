class Dai < Formula
  desc "AI-backed dependency upgrade advisor for package.json projects"
  homepage "https://github.com/HibiZA/dai"
  license "MIT"
  version "0.1.0" # Will be updated by automation

  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/HibiZA/dai/releases/download/v0.1.0/dai_darwin_arm64.tar.gz"
      sha256 "deadbeef" # Placeholder, will be updated by release automation
    else
      url "https://github.com/HibiZA/dai/releases/download/v0.1.0/dai_darwin_amd64.tar.gz"
      sha256 "deadbeef" # Placeholder, will be updated by release automation
    end
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/HibiZA/dai/releases/download/v0.1.0/dai_linux_amd64.tar.gz"
      sha256 "deadbeef" # Placeholder, will be updated by release automation
    else
      url "https://github.com/HibiZA/dai/releases/download/v0.1.0/dai_linux_arm64.tar.gz"
      sha256 "deadbeef" # Placeholder, will be updated by release automation
    end
  end

  def install
    bin.install "dai"
  end

  test do
    system "#{bin}/dai", "--version"
  end

  def caveats
    <<~EOS
      To use Dai CLI, you may need to configure your API keys:
      
      For OpenAI integration:
      $ export DAI_OPENAI_API_KEY=your_api_key
      
      For GitHub integration:
      $ export DAI_GITHUB_TOKEN=your_github_token
      
      You can add these to your shell profile for permanent use.
    EOS
  end
end 