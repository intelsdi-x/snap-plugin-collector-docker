# File managed by pluginsync

require_relative './spec_helper'
require 'specinfra/backend/docker_compose'

compose_yml = File.expand_path(File.join(__FILE__, "../docker-compose.yml"))
raise(Exception, "Missing docker-compose file: #{compose_yml}") unless File.exists? compose_yml

# NOTE: scan docker compose file and pull latest containers:
images = File.readlines(compose_yml).select {|l| l =~ /^\s*image:/}
images = images.collect{|l| l.split('image:').last.strip }.uniq
images.each do |i|
  puts `docker pull #{i}`
end

set :docker_compose_container, :snap

describe docker_compose(compose_yml) do

  # NOTE: If you need to wait for a service or create a database perform it in setup.rb
  setup = File.expand_path(File.join(__FILE__, '../setup.rb'))
  eval File.read setup if File.exists? setup

  its_container(:snap) do
    describe 'docker-compose.yml run' do
      TIMEOUT = 60

      describe "download Snap" do
        it {
          expect(cmd_with_retry("/opt/snap/bin/snaptel --version", :timeout => TIMEOUT).exit_status).to eq 0
          expect(cmd_with_retry("/opt/snap/sbin/snapteld --version", :timeout => TIMEOUT).exit_status).to eq 0
        }
      end

      if os[:family] == 'alpine'
        describe port(8181) do
          it { should be_listening }
        end
      end

      context "load Snap plugins" do
        describe command("snaptel plugin list") do
          it { load_all_plugins }
          its(:exit_status) { should eq 0 }
          its(:stdout) {
            plugins.each do |p|
              _ , name = p
              should contain(/#{name}/)
            end
          }
        end
      end

      describe file("/opt/snap/sbin/snapteld") do
        it { should be_file }
        it { should be_executable }
      end

      describe file("/opt/snap/bin/snaptel") do
        it { should be_file }
        it { should be_executable }
      end

      describe command("snapteld --version") do
        its(:exit_status) { should eq 0 }
        its(:stdout) { should contain(/#{ENV['SNAP_VERSION']}/) }
      end if ENV['SNAP_VERSION'] =~ /^\d+.\d+.\d+$/

      SnapUtils.tasks.each do |t|
        context "Snap task #{t}" do
          task_id = nil

          describe command("snaptel task create -t /plugin/examples/tasks/#{t}") do
            its(:exit_status) { should eq 0 }
            its(:stdout) { should contain(/Task created/) }
            it {
              id = subject.stdout.split("\n").find{|l|l=~/^ID:/}
              task_id = $1 if id.match(/^ID: (.*)$/)
              expect(task_id).to_not be_nil
              # NOTE we need a short pause before checking task state in case it fails:
              sleep 3
            }
          end

          describe command("snaptel task list") do
            its(:exit_status) { should eq 0 }
            its(:stdout) { should contain(/Running/) }
          end

          describe "Metrics in running tasks" do
            it {
              binding.pry if ENV["DEMO"] == "true"

              data = curl_json_api("http://127.0.0.1:8181/v1/tasks")
              task = data["body"]["ScheduledTasks"].find{|i| i['id'] == task_id}
              expect(task['id']).to eq task_id
              data = curl_json_api(task['href'])
              collect_metrics = data["body"]["workflow"]["collect"]["metrics"].collect{|k,v| k}

              config = load_yaml(SnapUtils.examples/"tasks/#{t}")
              config_metrics = config['workflow']['collect']['metrics'].collect{|k,v| k}

              config_metrics.each do |m|
                expect(collect_metrics).to include(m)
              end
            }
          end

          # NOTE: can not use the normal describe command(...) since we need to access task_id
          describe "Stop task" do
            it {
              c = command("snaptel task stop #{task_id}")
              expect(c.exit_status).to eq 0
              expect(c.stdout).to match(/Task stopped/)
            }
          end

          describe "Remove task" do
            it {
              c = command("snaptel task remove #{task_id}")
              expect(c.exit_status).to eq 0
              expect(c.stdout).to match(/Task removed/)
            }
          end
        end
      end
    end
  end

  # NOTE: If you need to perform additional checks such as database verification it be done at the end:
  verify = File.expand_path(File.join(__FILE__, '../verify.rb'))
  eval File.read verify if File.exists? verify
end
