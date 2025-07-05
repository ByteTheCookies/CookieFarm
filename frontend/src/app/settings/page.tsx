'use client';

import { useState, useEffect } from 'react';
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
import useSWR from 'swr';
import { Service, Protocol, ConfigData, SharedConfig } from '@/lib/types';
import { fetcher } from '@/lib/fetchers';
import {
  sharedConfigToConfigData,
  configDataToSharedConfig,
} from '@/lib/adapters';
import { toast } from 'sonner';
import { stringify } from 'yaml';

export default function Settings() {
  const [config, setConfig] = useState<ConfigData | null>(null);
  const [protocols, setProtocols] = useState<Protocol[]>([]);

  const { data: sharedConfig } = useSWR<SharedConfig>('/api/config', fetcher);
  const { data: fetchedProtocols } = useSWR<Protocol[]>(
    '/api/protocols',
    fetcher,
  );

  useEffect(() => {
    if (sharedConfig) {
      const configDataNew: ConfigData = sharedConfigToConfigData(sharedConfig);
      setConfig(configDataNew);
    }
  }, [sharedConfig]);

  useEffect(() => {
    if (fetchedProtocols) {
      setProtocols(fetchedProtocols);
    }
  }, [fetchedProtocols]);

  if (!config) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading configuration...</p>
      </div>
    );
  }

  const updateConfig = (
    section: keyof ConfigData,
    field: string,
    value: any, // eslint-disable-line
  ) => {
    setConfig(prev => ({
      ...prev!,
      [section]: {
        ...prev![section],
        [field]: value,
      },
    }));
  };

  const addService = () => {
    const newService: Service = {
      id: Date.now().toString(),
      name: '',
      port: 0,
    };
    setConfig(prev => ({
      ...prev!,
      services: [...prev!.services, newService],
    }));
  };

  const removeService = (id: string) => {
    setConfig(prev => ({
      ...prev!,
      services: prev!.services.filter(service => service.id !== id),
    }));
  };

  const updateService = (id: string, field: keyof Service, value: string) => {
    setConfig(prev => ({
      ...prev!,
      services: prev!.services.map(service =>
        service.id === id ? { ...service, [field]: value } : service,
      ),
    }));
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

  const saveConfiguration = async () => {
    try {
      const response = await fetch('/api/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ config: configDataToSharedConfig(config) }),
      });

      if (!response.ok) {
        throw new Error('Failed to save configuration');
      }
      toast.success('Configuration saved successfully');
    } catch (error) {
      toast.error('Error saving configuration');
      console.error('Error saving configuration:', error);
    }
  };

  const exportConfiguration = () => {
    const dataStr = stringify(configDataToSharedConfig(config), null, 2);
    const dataUri =
      'data:application/yaml;charset=utf-8,' + encodeURIComponent(dataStr);
    const exportFileDefaultName = 'config.yaml';

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
          <TabsTrigger value="services">Services</TabsTrigger>
          <TabsTrigger value="teams">Teams</TabsTrigger>
          <TabsTrigger value="general">General</TabsTrigger>
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
              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="url-flag-checker">Flag Checker URL</Label>
                  <InfoTooltip content="The URL of the flag checker server where flags will be submitted" />
                </div>
                <Input
                  id="url-flag-checker"
                  placeholder="http://example.com/flag-checker"
                  value={config.flagChecker.url_flag_checker}
                  onChange={e =>
                    updateConfig(
                      'flagChecker',
                      'url_flag_checker',
                      e.target.value,
                    )
                  }
                />
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
                  value={config.flagChecker.team_token}
                  onChange={e =>
                    updateConfig('flagChecker', 'team_token', e.target.value)
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
                    value={config.flagChecker.submit_flag_checker_time}
                    onChange={e =>
                      updateConfig(
                        'flagChecker',
                        'submit_flag_checker_time',
                        e.target.value,
                      )
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
                    value={config.flagChecker.max_flag_batch_size}
                    onChange={e =>
                      updateConfig(
                        'flagChecker',
                        'max_flag_batch_size',
                        e.target.value,
                      )
                    }
                  />
                </div>
              </div>

              <Separator />

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="regex-flag">Flag Regex</Label>
                  <InfoTooltip content="Regular expression used to validate flag format" />
                </div>
                <Input
                  id="regex-flag"
                  placeholder="^FLAG-[A-Za-z0-9]+$"
                  value={config.flagInfo.regex_flag}
                  onChange={e =>
                    updateConfig('flagInfo', 'regex_flag', e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="url-flag-ids">Flag IDs URL</Label>
                  <InfoTooltip content="URL to fetch flag IDs for validation and tracking" />
                </div>
                <Input
                  id="url-flag-ids"
                  placeholder="http://example.com/flag-ids"
                  value={config.flagInfo.url_flag_ids}
                  onChange={e =>
                    updateConfig('flagInfo', 'url_flag_ids', e.target.value)
                  }
                />
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
                      key={`${service.id}-${index}`}
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
              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="range-ip-teams">Range of IP Teams</Label>
                  <InfoTooltip content="Total number of teams participating in the CTF" />
                </div>
                <Input
                  id="range-ip-teams"
                  placeholder="10"
                  type="number"
                  value={config.teams.range_ip_teams}
                  onChange={e =>
                    updateConfig('teams', 'range_ip_teams', e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="my-team-id">Your Team ID</Label>
                  <InfoTooltip content="Your own team ID - this team will be excluded from attacks" />
                </div>
                <Input
                  id="my-team-id"
                  placeholder="5"
                  type="number"
                  value={config.teams.my_team_id}
                  onChange={e =>
                    updateConfig('teams', 'my_team_id', e.target.value)
                  }
                />
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
                  value={config.teams.nop_team}
                  onChange={e =>
                    updateConfig('teams', 'nop_team', e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="format-ip-teams">Team IP Format</Label>
                  <InfoTooltip content="IP address pattern for teams. Use {} as placeholder for team ID (e.g., 10.10.{}.1 becomes 10.10.5.1 for team 5)" />
                </div>
                <Input
                  id="format-ip-teams"
                  placeholder="10.10.{}.1"
                  value={config.teams.format_ip_teams}
                  onChange={e =>
                    updateConfig('teams', 'format_ip_teams', e.target.value)
                  }
                />
              </div>

              <div className="bg-muted mt-6 rounded-lg p-4">
                <h4 className="mb-2 font-semibold">Example Team IPs:</h4>
                <div className="text-muted-foreground space-y-1 text-sm">
                  <p>
                    Team 1: {config.teams.format_ip_teams.replace('{}', '1')}
                  </p>
                  <p>
                    Team 2: {config.teams.format_ip_teams.replace('{}', '2')}
                  </p>
                  <p>
                    Team 3: {config.teams.format_ip_teams.replace('{}', '3')}
                  </p>
                  <p>
                    Your Team ({config.teams.my_team_id}):{' '}
                    {config.teams.format_ip_teams.replace(
                      '{}',
                      config.teams.my_team_id.toString(),
                    )}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="general" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>General Configuration</CardTitle>
              <CardDescription>
                Configure general server settings and time parameters
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="protocol">Protocol</Label>
                  <InfoTooltip content="Protocol used by the server (e.g., HTTP, HTTPS)" />
                </div>
                <Select
                  value={config.general.protocol}
                  onValueChange={value =>
                    updateConfig('general', 'protocol', value)
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

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="tick-time">Tick Time (ms)</Label>
                  <InfoTooltip content="Interval in milliseconds between server ticks" />
                </div>
                <Input
                  id="tick-time"
                  placeholder="1000"
                  type="number"
                  value={config.general.tick_time}
                  onChange={e =>
                    updateConfig('general', 'tick_time', e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="flag-ttl">Flag TTL (seconds)</Label>
                  <InfoTooltip content="Time-to-live for flags before they expire" />
                </div>
                <Input
                  id="flag-ttl"
                  placeholder="300"
                  type="number"
                  value={config.general.flag_ttl}
                  onChange={e =>
                    updateConfig('general', 'flag_ttl', e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="start-time">Start Time</Label>
                  <InfoTooltip content="Start time of the CTF match (ISO 8601 format)" />
                </div>
                <Input
                  id="start-time"
                  placeholder="2023-01-01T00:00:00Z"
                  type="text"
                  value={config.general.start_time}
                  onChange={e =>
                    updateConfig('general', 'start_time', e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center">
                  <Label htmlFor="end-time">End Time</Label>
                  <InfoTooltip content="End time of the CTF match (ISO 8601 format)" />
                </div>
                <Input
                  id="end-time"
                  placeholder="2023-01-01T12:00:00Z"
                  type="text"
                  value={config.general.end_time}
                  onChange={e =>
                    updateConfig('general', 'end_time', e.target.value)
                  }
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
