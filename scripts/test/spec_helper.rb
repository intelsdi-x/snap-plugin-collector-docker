# File managed by pluginsync

require 'hashie'
require 'json'
require 'pathname'
require 'yaml'
require 'rspec/retry'
require 'dockerspec/serverspec'

begin
  require 'pry'
rescue LoadError
end

module SnapUtils
  def sh(arg)
    c = command(arg)
    puts c.stderr
    puts c.stdout
  end

  def build_path
    File.expand_path(File.join(__FILE__, '../../../build/linux/x86_64'))
  end

  def local_plugins
    Dir.chdir(build_path)
    @local_plugins ||= Dir.glob("snap-plugin-*")
  end

  def load_plugin(type, name, version="latest")
    plugin_name = "snap-plugin-#{type}-#{name}"

    # NOTE: use mock2 plugin when mock is requested, in general we should avoid mock collector.
    case name
    when 'mock' # TODO: revisit how we handle mock plugins
      url = "https://s3-us-west-2.amazonaws.com/snap.ci.snap-telemetry.io/snap/#{version}/linux/x86_64/#{plugin_name}2"
    when 'mock1', 'mock2', 'mock2-grpc', 'passthru', 'passthru-grpc', 'mock-file', 'mock-file-grpc'
      url = "https://s3-us-west-2.amazonaws.com/snap.ci.snap-telemetry.io/snap/#{version}/linux/x86_64/#{plugin_name}"
    else
      url = "https://s3-us-west-2.amazonaws.com/snap.ci.snap-telemetry.io/plugins/#{plugin_name}/#{version}/linux/x86_64/#{plugin_name}"
    end

    if local_plugins.include? plugin_name
      command("snaptel plugin load #{build_path}/#{plugin_name}").exit_status
    else
      command("curl -sfL #{url} -o /opt/snap/plugins/#{plugin_name}").exit_status
      command("snaptel plugin load /opt/snap/plugins/#{plugin_name}").exit_status
    end
  end

  def cmd_with_retry(arg, opt={ :timeout => 30 })
    cmd = command(arg)
    while cmd.exit_status != 0 or cmd.stdout == '' and opt[:timeout] > 0
      sleep 5
      opt[:timeout] -= 5
      cmd = command(arg)
    end
    return cmd
  end

  def curl_json_api(url)
    output = cmd_with_retry("curl #{url}").stdout
    if output.size > 0
      JSON.parse(output)
    else
      {}
    end
  end

  def load_json(file)
    file = File.expand_path file
    raise ArgumentError, "Invalid json file path: #{file}" unless File.exist? file
    JSON.parse(gsub_env(File.read file))
  end

  def load_yaml(file)
    file = File.expand_path file
    raise ArgumentError, "Invalid json file path: #{file}" unless File.exist? file
    YAML.load(gsub_env(File.read file))
  end

  def gsub_env(content)
    content.gsub(/\$([a-zA-Z_]+[a-zA-Z0-9_]*)|\$\{(.+)\}/) { ENV[$1 || $2] }
  end

  def self.examples
    Pathname.new(File.expand_path(File.join(__FILE__,'../../../examples')))
  end

  def self.tasks
    if ENV["TASK"] != ""
      pattern="#{examples}/tasks/#{ENV["TASK"]}"
    else
      pattern="#{examples}/tasks/*.y{a,}ml"
    end
    Dir.glob(pattern).collect{|f| File.basename f}
  end

  def add_plugins(plugins, type)
    plugins.flatten.compact.uniq.each do |name|
      @plugins.add([type, name])
    end
  end

  def parse_task(t)
    t.extend Hashie::Extensions::DeepFetch
    t.extend Hashie::Extensions::DeepFind

    m = t.deep_fetch("workflow", "collect", "metrics"){ |k| {} }
    collectors = m.collect do |k, v|
      k.match(/^\/intel\/(.*?)\/(.*?)/)
      # NOTE: procfs/* doesn't follow the convention, nor does disk/smart.
      if $1 == 'procfs' || $1 == 'disk'
        case $2
        when 'iface'
          'interface'
        when 'filesystem'
          'df'
        else
          $2
        end
      else
        $1
      end
    end
    add_plugins(collectors, 'collector')

    p = t.deep_find_all("process") || {}
    processors = p.collect do |i|
      if i.is_a? ::Array
        i.collect{|j| j["plugin_name"] if j.include? "plugin_name"}
      end
    end
    add_plugins(processors, 'processor')

    p = t.deep_find_all("publish") || {}
    publishers = p.collect do |i|
      if i.is_a? ::Array
        i.collect{|j| j["plugin_name"] if j.include? "plugin_name"}
      end
    end
    add_plugins(publishers, 'publisher')
  end

  def plugins
    @plugins ||= load_tasks
  end

  def load_tasks
    @plugins = Set.new
    SnapUtils.tasks.each do |t|
      y = load_yaml SnapUtils.examples/"tasks/#{t}"
      parse_task(y)
    end
    @plugins
  end

  def load_all_plugins
    plugins.each do |i|
      type, name = i
      load_plugin(type, name)
    end
  end
end

RSpec.configure do |c|
  c.formatter = 'documentation'
  c.mock_framework = :rspec
  c.verbose_retry = true
  c.order = 'default'
  c.include SnapUtils
  if ENV["DEMO"] == "true" then
    Pry.config.pager = false

    Pry.hooks.add_hook(:before_session, "notice") do |output, binding, pry|
      output.puts "Setup complete for DEMO mode. When you are finished checking out Snap please type 'exit-program' to shutdown containers."
    end
  end
end
