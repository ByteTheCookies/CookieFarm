'use client';

import { useState } from 'react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Separator } from '@/components/ui/separator';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Info, Plus, Trash2, Save, Download } from 'lucide-react';

interface Service {
  id: string;
  name: string;
  port: string;
}

interface ConfigData {
  flagChecker: {
    ipAddress: string;
    port: string;
    teamToken: string;
    submitTime: string;
    batchSize: string;
    protocol: string;
  };
  rounds: {
    duration: string;
    startTime: string;
    endTime: string;
  };
  services: Service[];
  teams: {
    numberOfTeams: string;
    nopTeamId: string;
    ownTeamId: string;
    ipFormat: string;
  };
  notifications: {
    flagSubmissionAlerts: boolean;
    serviceDownAlerts: boolean;
    roundChangeNotifications: boolean;
    exploitSuccessAlerts: boolean;
  };
  security: {
    apiKey: string;
    sessionTimeout: string;
    twoFactorAuth: boolean;
    auditLogging: boolean;
  };
}

const protocols = [
  { value: 'cc_http', label: 'CC HTTP' },
  { value: 'cc_grpc', label: 'CC gRPC' },
  { value: 'defcon_http', label: 'DEFCON HTTP' },
  { value: 'custom_tcp', label: 'Custom TCP' },
  { value: 'custom_udp', label: 'Custom UDP' },
];

export default function Settings() {
  const [config, setConfig] = useState<ConfigData>({
    flagChecker: {
      ipAddress: '192.168.1.100',
      port: '8080',
      teamToken: '',
      submitTime: '30',
      batchSize: '100',
      protocol: 'cc_http',
    },
    rounds: {
      duration: '120',
      startTime: '',
      endTime: '',
    },
    services: [
      { id: '1', name: 'web', port: '80' },
      { id: '2', name: 'database', port: '3306' },
    ],
    teams: {
      numberOfTeams: '10',
      nopTeamId: '1',
      ownTeamId: '5',
      ipFormat: '10.10.{}.1',
    },
    notifications: {
      flagSubmissionAlerts: true,
      serviceDownAlerts: true,
      roundChangeNotifications: false,
      exploitSuccessAlerts: true,
    },
    security: {
      apiKey: '',
      sessionTimeout: '8',
      twoFactorAuth: false,
      auditLogging: true,
    },
  });

  const updateConfig = (
    section: keyof ConfigData,
    field: string,
    value: any, // eslint-disable-line
  ) => {
    setConfig(prev => ({
      ...prev,
      [section]: {
        ...prev[section],
        [field]: value,
      },
    }));
  };

  const addService = () => {
    const newService: Service = {
      id: Date.now().toString(),
      name: '',
      port: '',
    };
    setConfig(prev => ({
      ...prev,
      services: [...prev.services, newService],
    }));
  };

  const removeService = (id: string) => {
    setConfig(prev => ({
      ...prev,
      services: prev.services.filter(service => service.id !== id),
    }));
  };

  const updateService = (id: string, field: keyof Service, value: string) => {
    setConfig(prev => ({
      ...prev,
      services: prev.services.map(service =>
        service.id === id ? { ...service, [field]: value } : service,
      ),
    }));
  };

  const calculateTotalTicks = () => {
    const { startTime, endTime, duration } = config.rounds;
    if (!startTime || !endTime || !duration) return 0;

    const start = new Date(startTime).getTime();
    const end = new Date(endTime).getTime();
    const durationMs = Number.parseInt(duration) * 1000;

    if (end <= start || durationMs <= 0) return 0;

    return Math.floor((end - start) / durationMs);
  };

  const InfoTooltip = ({ content }: { content: string }) => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Info className="text-muted-foreground ml-2 h-4 w-4 cursor-help" />
        </TooltipTrigger>
        <TooltipContent className="max-w-xs">
          <p>{content}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );

  const saveConfiguration = () => {
    // Implementation for saving configuration
    console.log('Saving configuration:', config);
  };

  const exportConfiguration = () => {
    const dataStr = JSON.stringify(config, null, 2);
    const dataUri =
      'data:application/json;charset=utf-8,' + encodeURIComponent(dataStr);
    const exportFileDefaultName = 'cookiefarm-config.json';

    const linkElement = document.createElement('a');
    linkElement.setAttribute('href', dataUri);
    linkElement.setAttribute('download', exportFileDefaultName);
    linkElement.click();
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold">Settings</h1>
          <p className="text-muted-foreground">
            Configure CookieFarm settings and server parameters
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={exportConfiguration}>
            <Download className="mr-2 h-4 w-4" />
            Export Config
          </Button>
          <Button onClick={saveConfiguration}>
            <Save className="mr-2 h-4 w-4" />
            Save Changes
          </Button>
        </div>
      </div>

      <Tabs defaultValue="flag-checker" className="space-y-6">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="flag-checker">Flag Checker</TabsTrigger>
          <TabsTrigger value="rounds">Rounds</TabsTrigger>
          <TabsTrigger value="services">Services</TabsTrigger>
          <TabsTrigger value="teams">Teams</TabsTrigger>
          <TabsTrigger value="system">System</TabsTrigger>
        </TabsList>

        <TabsContent value="flag-checker" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Flag Checker Configuration</CardTitle>
              <CardDescription>
                Configure flag submission settings and communication parameters
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <div className="flex items-center">
                    <Label htmlFor="ip-address">IP Address</Label>
                    <InfoTooltip content="The IP address of the flag checker server where flags will be submitted" />
                  </div>
                  <Input
                    id="ip-address"
                    placeholder="192.168.1.100"
                    value={config.flagChecker.ipAddress}
                    onChange={e =>
                      updateConfig('flagChecker', 'ipAddress', e.target.value)
                    }
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center">
                    <Label htmlFor="port">Port</Label>
                    <InfoTooltip content="The port number on which the flag checker service is running" />
                  </div>
                  <Input
                    id="port"
                    placeholder="8080"
                    type="number"
                    value={config.flagChecker.port}
                    onChange={e =>
                      updateConfig('flagChecker', 'port', e.target.value)
                    }
                  />
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="team-token">Team Token</Label>
                  <InfoTooltip content="Authentication token provided by the CTF organizers for your team" />
                </div>
                <Input
                  id="team-token"
                  placeholder="your-team-token-here"
                  type="password"
                  value={config.flagChecker.teamToken}
                  onChange={e =>
                    updateConfig('flagChecker', 'teamToken', e.target.value)
                  }
                />
              </div>

              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <div className="flex items-center">
                    <Label htmlFor="submit-time">Submit Time (seconds)</Label>
                    <InfoTooltip content="How many seconds to wait before submitting all pending flags in batch" />
                  </div>
                  <Input
                    id="submit-time"
                    placeholder="30"
                    type="number"
                    value={config.flagChecker.submitTime}
                    onChange={e =>
                      updateConfig('flagChecker', 'submitTime', e.target.value)
                    }
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center">
                    <Label htmlFor="batch-size">Batch Size</Label>
                    <InfoTooltip content="Maximum number of flags to submit in a single batch request" />
                  </div>
                  <Input
                    id="batch-size"
                    placeholder="100"
                    type="number"
                    value={config.flagChecker.batchSize}
                    onChange={e =>
                      updateConfig('flagChecker', 'batchSize', e.target.value)
                    }
                  />
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="protocol">Communication Protocol</Label>
                  <InfoTooltip content="The protocol used to communicate with the flag checker (depends on CTF platform)" />
                </div>
                <Select
                  value={config.flagChecker.protocol}
                  onValueChange={value =>
                    updateConfig('flagChecker', 'protocol', value)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select protocol" />
                  </SelectTrigger>
                  <SelectContent>
                    {protocols.map(protocol => (
                      <SelectItem key={protocol.value} value={protocol.value}>
                        {protocol.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="rounds" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Rounds Configuration</CardTitle>
              <CardDescription>
                Set up match timing, round duration, and schedule
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="round-duration">
                    Round Duration (seconds)
                  </Label>
                  <InfoTooltip content="Duration of each round/tick in seconds. Common values are 60, 120, or 300 seconds" />
                </div>
                <Input
                  id="round-duration"
                  placeholder="120"
                  type="number"
                  value={config.rounds.duration}
                  onChange={e =>
                    updateConfig('rounds', 'duration', e.target.value)
                  }
                />
              </div>

              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <div className="flex items-center">
                    <Label htmlFor="start-time">Match Start Time</Label>
                    <InfoTooltip content="When the CTF match begins. Use local time format" />
                  </div>
                  <Input
                    id="start-time"
                    type="datetime-local"
                    value={config.rounds.startTime}
                    onChange={e =>
                      updateConfig('rounds', 'startTime', e.target.value)
                    }
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center">
                    <Label htmlFor="end-time">Match End Time</Label>
                    <InfoTooltip content="When the CTF match ends. Use local time format" />
                  </div>
                  <Input
                    id="end-time"
                    type="datetime-local"
                    value={config.rounds.endTime}
                    onChange={e =>
                      updateConfig('rounds', 'endTime', e.target.value)
                    }
                  />
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label>Total Number of Ticks</Label>
                  <InfoTooltip content="Automatically calculated based on match duration and round length" />
                </div>
                <div className="bg-muted rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <span className="text-primary text-3xl font-bold">
                        {calculateTotalTicks()}
                      </span>
                      <span className="text-muted-foreground ml-2 text-lg">
                        ticks
                      </span>
                    </div>
                    <Badge variant="outline">
                      {config.rounds.duration}s per round
                    </Badge>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="services" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Services Configuration</CardTitle>
                  <CardDescription>
                    Manage services that will be monitored during the CTF
                  </CardDescription>
                </div>
                <Button onClick={addService} size="sm">
                  <Plus className="mr-2 h-4 w-4" />
                  Add Service
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {config.services.length === 0 ? (
                <div className="text-muted-foreground py-8 text-center">
                  <p>No services configured yet.</p>
                  <p className="text-sm">
                    Click &quot; Add Service &quot; to get started.
                  </p>
                </div>
              ) : (
                <div className="space-y-4">
                  {config.services.map((service, index) => (
                    <div
                      key={service.id}
                      className="flex items-center gap-4 rounded-lg border p-4"
                    >
                      <Badge variant="outline">Service {index + 1}</Badge>
                      <div className="grid flex-1 grid-cols-1 gap-4 md:grid-cols-2">
                        <div className="space-y-2">
                          <div className="flex items-center">
                            <Label>Service Name</Label>
                            <InfoTooltip content="Name of the service (e.g., web, database, api)" />
                          </div>
                          <Input
                            placeholder="web"
                            value={service.name}
                            onChange={e =>
                              updateService(service.id, 'name', e.target.value)
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <div className="flex items-center">
                            <Label>Service Port</Label>
                            <InfoTooltip content="Port number on which this service runs on team machines" />
                          </div>
                          <Input
                            placeholder="80"
                            type="number"
                            value={service.port}
                            onChange={e =>
                              updateService(service.id, 'port', e.target.value)
                            }
                          />
                        </div>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => removeService(service.id)}
                        className="text-destructive hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="teams" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Teams Configuration</CardTitle>
              <CardDescription>
                Configure team settings and network parameters
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <div className="flex items-center">
                    <Label htmlFor="num-teams">Number of Teams</Label>
                    <InfoTooltip content="Total number of teams participating in the CTF" />
                  </div>
                  <Input
                    id="num-teams"
                    placeholder="10"
                    type="number"
                    value={config.teams.numberOfTeams}
                    onChange={e =>
                      updateConfig('teams', 'numberOfTeams', e.target.value)
                    }
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center">
                    <Label htmlFor="own-team">Your Team ID</Label>
                    <InfoTooltip content="Your own team ID - this team will be excluded from attacks" />
                  </div>
                  <Input
                    id="own-team"
                    placeholder="5"
                    type="number"
                    value={config.teams.ownTeamId}
                    onChange={e =>
                      updateConfig('teams', 'ownTeamId', e.target.value)
                    }
                  />
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="nop-team">NOP Team ID</Label>
                  <InfoTooltip content="ID of the NOP (No Operation) team - usually used for testing or as a baseline" />
                </div>
                <Input
                  id="nop-team"
                  placeholder="1"
                  type="number"
                  value={config.teams.nopTeamId}
                  onChange={e =>
                    updateConfig('teams', 'nopTeamId', e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="ip-format">Team IP Format</Label>
                  <InfoTooltip content="IP address pattern for teams. Use {} as placeholder for team ID (e.g., 10.10.{}.1 becomes 10.10.5.1 for team 5)" />
                </div>
                <Input
                  id="ip-format"
                  placeholder="10.10.{}.1"
                  value={config.teams.ipFormat}
                  onChange={e =>
                    updateConfig('teams', 'ipFormat', e.target.value)
                  }
                />
              </div>

              <div className="bg-muted mt-6 rounded-lg p-4">
                <h4 className="mb-2 font-semibold">Example Team IPs:</h4>
                <div className="text-muted-foreground space-y-1 text-sm">
                  <p>Team 1: {config.teams.ipFormat.replace('{}', '1')}</p>
                  <p>Team 2: {config.teams.ipFormat.replace('{}', '2')}</p>
                  <p>Team 3: {config.teams.ipFormat.replace('{}', '3')}</p>
                  <p>
                    Your Team ({config.teams.ownTeamId}):{' '}
                    {config.teams.ipFormat.replace(
                      '{}',
                      config.teams.ownTeamId,
                    )}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="system" className="space-y-6">
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Notification Settings */}
            <Card>
              <CardHeader>
                <CardTitle>Notifications</CardTitle>
                <CardDescription>
                  Configure alert and notification preferences
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Flag Submission Alerts</Label>
                    <p className="text-muted-foreground text-sm">
                      Get notified when flags are submitted
                    </p>
                  </div>
                  <Switch
                    checked={config.notifications.flagSubmissionAlerts}
                    onCheckedChange={checked =>
                      updateConfig(
                        'notifications',
                        'flagSubmissionAlerts',
                        checked,
                      )
                    }
                  />
                </div>
                <Separator />
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Service Down Alerts</Label>
                    <p className="text-muted-foreground text-sm">
                      Alert when services go offline
                    </p>
                  </div>
                  <Switch
                    checked={config.notifications.serviceDownAlerts}
                    onCheckedChange={checked =>
                      updateConfig(
                        'notifications',
                        'serviceDownAlerts',
                        checked,
                      )
                    }
                  />
                </div>
                <Separator />
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Round Change Notifications</Label>
                    <p className="text-muted-foreground text-sm">
                      Notify when rounds change
                    </p>
                  </div>
                  <Switch
                    checked={config.notifications.roundChangeNotifications}
                    onCheckedChange={checked =>
                      updateConfig(
                        'notifications',
                        'roundChangeNotifications',
                        checked,
                      )
                    }
                  />
                </div>
                <Separator />
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Exploit Success Alerts</Label>
                    <p className="text-muted-foreground text-sm">
                      Alert on successful exploits
                    </p>
                  </div>
                  <Switch
                    checked={config.notifications.exploitSuccessAlerts}
                    onCheckedChange={checked =>
                      updateConfig(
                        'notifications',
                        'exploitSuccessAlerts',
                        checked,
                      )
                    }
                  />
                </div>
              </CardContent>
            </Card>

            {/* Security Settings */}
            <Card>
              <CardHeader>
                <CardTitle>Security</CardTitle>
                <CardDescription>
                  Configure security and access control settings
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="api-key">API Key</Label>
                  <Input
                    id="api-key"
                    type="password"
                    placeholder="••••••••••••••••"
                    value={config.security.apiKey}
                    onChange={e =>
                      updateConfig('security', 'apiKey', e.target.value)
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="session-timeout">
                    Session Timeout (hours)
                  </Label>
                  <Input
                    id="session-timeout"
                    type="number"
                    placeholder="8"
                    value={config.security.sessionTimeout}
                    onChange={e =>
                      updateConfig('security', 'sessionTimeout', e.target.value)
                    }
                  />
                </div>
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Two-Factor Authentication</Label>
                    <p className="text-muted-foreground text-sm">
                      Enable 2FA for additional security
                    </p>
                  </div>
                  <Switch
                    checked={config.security.twoFactorAuth}
                    onCheckedChange={checked =>
                      updateConfig('security', 'twoFactorAuth', checked)
                    }
                  />
                </div>
                <Separator />
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Audit Logging</Label>
                    <p className="text-muted-foreground text-sm">
                      Log all system activities
                    </p>
                  </div>
                  <Switch
                    checked={config.security.auditLogging}
                    onCheckedChange={checked =>
                      updateConfig('security', 'auditLogging', checked)
                    }
                  />
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
