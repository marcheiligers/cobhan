# frozen_string_literal: true

require_relative "lib/cobhan-demo-lib/version"

Gem::Specification.new do |spec|
  spec.name          = "cobhan-demo-lib"
  spec.version       = CobhanDemoLib::VERSION
  spec.authors       = ["Marc Heiligers"]
  spec.email         = ["mheiligers@godaddy.com"]

  spec.summary       = "Test library for Cobhan Ruby"
  spec.homepage      = "https://github.com/jgowdy/cobhan"
  spec.license       = "MIT"
  spec.required_ruby_version = ">= 2.4.0"

  spec.metadata["homepage_uri"] = spec.homepage
  spec.metadata["source_code_uri"] = spec.homepage

  # Specify which files should be added to the gem when it is released.
  # The `git ls-files -z` loads the files in the RubyGem that have been added into git.
  spec.files = Dir.chdir(File.expand_path(__dir__)) do
    `git ls-files -z`.split("\x0").reject { |f| f.match(%r{\A(?:test|spec|features)/}) }
  end
  spec.bindir        = "exe"
  spec.executables   = spec.files.grep(%r{\Aexe/}) { |f| File.basename(f) }
  spec.require_paths = ["lib"]

  # Uncomment to register a new dependency of your gem
  spec.add_dependency "cobhan"

  # For more information and examples about making a new gem, checkout our
  # guide at: https://bundler.io/guides/creating_gem.html
end
