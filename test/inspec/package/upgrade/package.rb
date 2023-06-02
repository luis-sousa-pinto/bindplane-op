version = input('version', value: '')

describe package('bindplane') do
    it { should be_installed }
    its('version') { should eq version }
end

describe systemd_service('bindplane') do
    it { should be_installed }
    it { should be_enabled }
    it { should be_running }
end
