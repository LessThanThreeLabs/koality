basePath = '..';

files = [
  JASMINE,
  JASMINE_ADAPTER,
  ANGULAR_SCENARIO,
  ANGULAR_SCENARIO_ADAPTER,
  'test/e2e/*.coffee'
];

reporters = ['progress'];

port = 8080;

runnerPort = 9100;

colors = true;

logLevel = LOG_INFO;

autoWatch = true;

browsers = ['Chrome', 'Firefox', 'Safari', 'PhantomJS'];

captureTimeout = 5000;

singleRun = false;

proxies = {
  '/': 'https://localhost:10443/'
};

urlRoot = '/karma';