describe package('bindplane') do
    it { should_not be_installed }
end

describe file('/usr/local/bin/bindplane') do
    it { should_not exist }
end

if os.family == 'debian'
    # Uninstall should not remove a modified config file.
    describe file('/etc/bindplane/config.yaml') do
        it { should exist }
    end
else
    # Uninstall on rhel platforms preserves the config file.
    describe file('/etc/bindplane/config.yaml.rpmsave') do
        it { should exist }
    end
end

# Uninstall should not remove the database file.
describe file('/var/lib/bindplane/storage/bindplane.db') do
    it { should exist }
end

# Uninstall should not remove the log file.
describe file('/var/log/bindplane/bindplane.log') do
    it { should exist }
end

describe file('/usr/lib/systemd/system/bindplane.service') do
    it { should_not exist }
end

# Uninstall should not remove user.
describe user('bindplane') do
    it { should exist }
end

# Uninstall should ot remove group.
describe group('bindplane') do
    it { should exist }
end

describe systemd_service('bindplane') do
    it { should_not be_installed }
    it { should_not be_enabled }
    it { should_not be_running }
end

describe port(3001) do
    it { should_not be_listening }
end

describe processes('bindplane') do
    it { should_not exist }
end
