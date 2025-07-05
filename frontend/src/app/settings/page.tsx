'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Settings, Save, RefreshCw, Plus, Trash2, Edit3 } from 'lucide-react';
import { z } from 'zod';
import useSWR from 'swr';
import axios from 'axios';

// Types based on your stepper component
interface Service {
  id: string;
  name: string;
  port: number;
}

interface ConfigData {
  general: {
    protocol: string;
    tick_time: number;
    flag_ttl: number;
    start_time: string;
    end_time: string;
  };
  flagChecker: {
    url_flag_checker: string;
    team_token: string;
    submit_flag_checker_time: number;
    max_flag_batch_size: number;
  };
  flagInfo: {
    regex_flag: string;
    url_flag_ids: string;
  };
  services: Service[];
  teams: {
    range_ip_teams: number;
    nop_team: number;
    my_team_id: number;
    format_ip_teams: string;
  };
}

interface Protocol {
  value: string;
  label: string;
}

// Validation schemas
const serviceSchema = z.object({
  id: z.string(),
  name: z.string().min(1, 'Service name is required'),
  port: z.number().min(1, 'Port must be greater than 0'),
});

const generalSchema = z.object({
  protocol: z.string().min(1, 'Protocol is required'),
  tick_time: z.number().min(1, 'Tick time must be greater than 0'),
  flag_ttl: z.number().min(0, 'Flag time to live must be positive'),
  start_time: z.string().min(1, 'Start time is required'),
  end_time: z.string().min(1, 'End time is required'),
});

const flagCheckerSchema = z.object({
  url_flag_checker: z
    .string()
    .min(1, 'URL is required')
    .url('Must be a valid URL'),
  team_token: z.string().min(1, 'Team token is required'),
  submit_flag_checker_time: z
    .number()
    .min(1, 'Submit time must be greater than 0'),
  max_flag_batch_size: z.number().min(1, 'Batch size must be greater than 0'),
});

const flagInfoSchema = z.object({
  regex_flag: z.string().min(1, 'Flag format is required'),
  url_flag_ids: z
    .string()
    .min(1, 'Flag IDs URL is required')
    .url('Must be a valid URL'),
});

const teamsSchema = z.object({
  range_ip_teams: z.number().min(1, 'Number of teams must be greater than 0'),
  nop_team: z.number().min(0, 'NOP Team ID must be valid'),
  my_team_id: z.number().min(1, 'Own Team ID must be greater than 0'),
  format_ip_teams: z
    .string()
    .min(1, 'IP Format is required')
    .regex(/\{\}/, 'IP Format must contain {}'),
});

const fetcher = (url: string) => fetch(url).then(res => res.json());

export default function SettingsPage() {
  const [config, setConfig] = useState<ConfigData | null>(null);
  const [editingSection, setEditingSection] = useState<string | null>(null);
  const [tempConfig, setTempConfig] = useState<ConfigData | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  const { data: protocols = [] } = useSWR<Protocol[]>(
    '/api/protocols',
    fetcher,
  );

  // Load current configuration
  useEffect(() => {
    const loadConfig = async () => {
      try {
        const response = await fetch('/api/config');

        // Check if the response is ok and contains JSON
        if (!response.ok) {
          if (response.status === 404) {
            console.log('Configuration API not found - using default state');
            setIsLoading(false);
            return;
          }
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        // Check if response is JSON
        const contentType = response.headers.get('content-type');
        if (!contentType || !contentType.includes('application/json')) {
          console.log('API response is not JSON - using default state');
          setIsLoading(false);
          return;
        }

        const data = await response.json();

        // Check if data has the expected structure
        if (data && data.config) {
          setConfig(data.config);
        } else {
          console.log('Invalid configuration structure received');
        }
      } catch (error) {
        console.error('Failed to load configuration:', error);
        // Don't throw the error, just log it and continue with no config
      } finally {
        setIsLoading(false);
      }
    };
    loadConfig();
  }, []);

  const startEditing = (section: string) => {
    setEditingSection(section);
    setTempConfig(config);
    setErrors({});
  };

  const cancelEditing = () => {
    setEditingSection(null);
    setTempConfig(null);
    setErrors({});
  };

  const updateTempConfig = (
    section: keyof ConfigData,
    field: string,
    value: any,
  ) => {
    if (!tempConfig) return;

    let processedValue = value;

    // Convert string numbers to actual numbers for numeric fields
    if (section === 'general' && ['tick_time', 'flag_ttl'].includes(field)) {
      processedValue = value === '' ? 0 : Number.parseInt(value) || 0;
    }
    if (
      section === 'flagChecker' &&
      ['submit_flag_checker_time', 'max_flag_batch_size'].includes(field)
    ) {
      processedValue = value === '' ? 0 : Number.parseInt(value) || 0;
    }
    if (
      section === 'teams' &&
      ['range_ip_teams', 'nop_team', 'my_team_id'].includes(field)
    ) {
      processedValue = value === '' ? 0 : Number.parseInt(value) || 0;
    }

    setTempConfig(prev => ({
      ...prev!,
      [section]: {
        ...prev![section],
        [field]: processedValue,
      },
    }));

    // Clear errors for this field
    setErrors(prev => {
      const newErrors = { ...prev };
      delete newErrors[`${section}.${field}`];
      return newErrors;
    });
  };

  const addService = () => {
    if (!tempConfig) return;
    const newService: Service = {
      id: Date.now().toString(),
      name: '',
      port: 0,
    };
    setTempConfig(prev => ({
      ...prev!,
      services: [...prev!.services, newService],
    }));
  };

  const removeService = (id: string) => {
    if (!tempConfig) return;
    setTempConfig(prev => ({
      ...prev!,
      services: prev!.services.filter(service => service.id !== id),
    }));
  };

  const updateService = (
    id: string,
    field: keyof Service,
    value: string | number,
  ) => {
    if (!tempConfig) return;
    let processedValue = value;

    if (field === 'port') {
      processedValue =
        typeof value === 'string'
          ? value === ''
            ? 0
            : Number.parseInt(value) || 0
          : value;
    }

    setTempConfig(prev => ({
      ...prev!,
      services: prev!.services.map(service =>
        service.id === id ? { ...service, [field]: processedValue } : service,
      ),
    }));

    // Clear errors for this field
    setErrors(prev => {
      const newErrors = { ...prev };
      delete newErrors[`services.${id}.${field}`];
      return newErrors;
    });
  };

  const validateSection = (section: string): boolean => {
    if (!tempConfig) return false;

    const sectionErrors: Record<string, string> = {};

    if (section === 'general') {
      const result = generalSchema.safeParse(tempConfig.general);
      if (!result.success) {
        for (const issue of result.error.issues) {
          sectionErrors[`general.${issue.path[0]}`] = issue.message;
        }
      }
    } else if (section === 'flagChecker') {
      const result = flagCheckerSchema.safeParse(tempConfig.flagChecker);
      if (!result.success) {
        for (const issue of result.error.issues) {
          sectionErrors[`flagChecker.${issue.path[0]}`] = issue.message;
        }
      }
    } else if (section === 'flagInfo') {
      const result = flagInfoSchema.safeParse(tempConfig.flagInfo);
      if (!result.success) {
        for (const issue of result.error.issues) {
          sectionErrors[`flagInfo.${issue.path[0]}`] = issue.message;
        }
      }
    } else if (section === 'services') {
      if (tempConfig.services.length === 0) {
        sectionErrors['services'] = 'At least one service is required';
      } else {
        tempConfig.services.forEach(service => {
          const result = serviceSchema.safeParse(service);
          if (!result.success) {
            for (const issue of result.error.issues) {
              sectionErrors[`services.${service.id}.${issue.path[0]}`] =
                issue.message;
            }
          }
        });
      }
    } else if (section === 'teams') {
      const result = teamsSchema.safeParse(tempConfig.teams);
      if (!result.success) {
        for (const issue of result.error.issues) {
          sectionErrors[`teams.${issue.path[0]}`] = issue.message;
        }
      }
    }

    setErrors(sectionErrors);
    return Object.keys(sectionErrors).length === 0;
  };

  const saveSection = async () => {
    if (!tempConfig || !editingSection) return;

    if (!validateSection(editingSection)) {
      return;
    }

    setIsSaving(true);
    try {
      const configFormatted = {
        config: {
          server: {
            ...tempConfig.general,
            ...tempConfig.flagChecker,
          },
          client: {
            services: tempConfig.services.map(({ id, ...service }) => service),
            ...tempConfig.teams,
            ...tempConfig.flagInfo,
          },
          configured: true,
        },
      };

      await axios.post('/api/config', configFormatted, {
        headers: { 'Content-Type': 'application/json' },
      });

      setConfig(tempConfig);
      setEditingSection(null);
      setTempConfig(null);
      setErrors({});
    } catch (error) {
      console.error('Failed to save configuration:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const calculateTotalTicks = (configData: ConfigData) => {
    const { start_time, end_time, tick_time } = configData.general;
    if (!start_time || !end_time || !tick_time) return 0;

    const start = new Date(start_time).getTime();
    const end = new Date(end_time).getTime();
    const tickMs = tick_time * 1000;

    if (end <= start || tickMs <= 0) return 0;
    return Math.floor((end - start) / tickMs);
  };

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (!config) {
    return (
      <div className="bg-background min-h-screen p-6">
        <div className="mx-auto max-w-4xl">
          <div className="flex min-h-[60vh] items-center justify-center">
            <Card className="w-96">
              <CardHeader>
                <CardTitle>No Configuration Found</CardTitle>
                <CardDescription>
                  No saved configuration was found. You can either:
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <p className="text-muted-foreground text-sm">
                    • Complete the initial server configuration setup
                  </p>
                  <p className="text-muted-foreground text-sm">
                    • Or create a new configuration manually
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button
                    onClick={() => (window.location.href = '/setup')}
                    className="flex-1"
                  >
                    Go to Setup
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => {
                      // Create a minimal default config for demonstration
                      const defaultConfig: ConfigData = {
                        general: {
                          protocol: '',
                          tick_time: 120,
                          flag_ttl: 5,
                          start_time: new Date().toISOString().slice(0, 16),
                          end_time: new Date(Date.now() + 8 * 60 * 60 * 1000)
                            .toISOString()
                            .slice(0, 16),
                        },
                        flagChecker: {
                          url_flag_checker: '',
                          team_token: '',
                          submit_flag_checker_time: 120,
                          max_flag_batch_size: 1000,
                        },
                        flagInfo: {
                          regex_flag: '[A-Z0-9]{31}=',
                          url_flag_ids: '',
                        },
                        services: [],
                        teams: {
                          range_ip_teams: 10,
                          nop_team: 0,
                          my_team_id: 1,
                          format_ip_teams: '10.10.{}.1',
                        },
                      };
                      setConfig(defaultConfig);
                    }}
                    className="flex-1"
                  >
                    Create New
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-background min-h-screen p-6">
      <div className="mx-auto max-w-4xl space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Settings className="h-8 w-8" />
            <div>
              <h1 className="text-3xl font-bold">Settings</h1>
              <p className="text-muted-foreground">
                Manage your server configuration
              </p>
            </div>
          </div>
          <Badge variant="secondary" className="text-sm">
            Configured
          </Badge>
        </div>

        <Tabs defaultValue="general" className="space-y-6">
          <TabsList className="grid w-full grid-cols-5">
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="flagChecker">Flag Checker</TabsTrigger>
            <TabsTrigger value="flagInfo">Flag Info</TabsTrigger>
            <TabsTrigger value="services">Services</TabsTrigger>
            <TabsTrigger value="teams">Teams</TabsTrigger>
          </TabsList>

          <TabsContent value="general">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>General Settings</CardTitle>
                  <CardDescription>
                    Basic server configuration parameters
                  </CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => startEditing('general')}
                  disabled={editingSection === 'general'}
                >
                  <Edit3 className="mr-2 h-4 w-4" />
                  Edit
                </Button>
              </CardHeader>
              <CardContent className="space-y-4">
                {editingSection === 'general' && tempConfig ? (
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label>Protocol</Label>
                        <Select
                          value={tempConfig.general.protocol}
                          onValueChange={value =>
                            updateTempConfig('general', 'protocol', value)
                          }
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {protocols.map(protocol => (
                              <SelectItem
                                key={protocol.value}
                                value={protocol.value}
                              >
                                {protocol.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        {errors['general.protocol'] && (
                          <p className="text-destructive text-xs">
                            {errors['general.protocol']}
                          </p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label>Tick Time (seconds)</Label>
                        <Input
                          type="number"
                          value={tempConfig.general.tick_time || ''}
                          onChange={e =>
                            updateTempConfig(
                              'general',
                              'tick_time',
                              e.target.value,
                            )
                          }
                        />
                        {errors['general.tick_time'] && (
                          <p className="text-destructive text-xs">
                            {errors['general.tick_time']}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label>Flag TTL (rounds)</Label>
                        <Input
                          type="number"
                          value={tempConfig.general.flag_ttl || ''}
                          onChange={e =>
                            updateTempConfig(
                              'general',
                              'flag_ttl',
                              e.target.value,
                            )
                          }
                        />
                        {errors['general.flag_ttl'] && (
                          <p className="text-destructive text-xs">
                            {errors['general.flag_ttl']}
                          </p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label>Total Ticks</Label>
                        <div className="bg-muted rounded-md p-2 text-center">
                          <span className="text-primary text-lg font-bold">
                            {calculateTotalTicks(tempConfig)}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label>Start Time</Label>
                        <Input
                          type="datetime-local"
                          value={tempConfig.general.start_time}
                          onChange={e =>
                            updateTempConfig(
                              'general',
                              'start_time',
                              e.target.value,
                            )
                          }
                        />
                        {errors['general.start_time'] && (
                          <p className="text-destructive text-xs">
                            {errors['general.start_time']}
                          </p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label>End Time</Label>
                        <Input
                          type="datetime-local"
                          value={tempConfig.general.end_time}
                          onChange={e =>
                            updateTempConfig(
                              'general',
                              'end_time',
                              e.target.value,
                            )
                          }
                        />
                        {errors['general.end_time'] && (
                          <p className="text-destructive text-xs">
                            {errors['general.end_time']}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="flex gap-2 pt-4">
                      <Button onClick={saveSection} disabled={isSaving}>
                        <Save className="mr-2 h-4 w-4" />
                        {isSaving ? 'Saving...' : 'Save Changes'}
                      </Button>
                      <Button variant="outline" onClick={cancelEditing}>
                        Cancel
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="grid grid-cols-2 gap-6">
                    <div className="space-y-3">
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Protocol
                        </Label>
                        <p className="font-medium">{config.general.protocol}</p>
                      </div>
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Tick Time
                        </Label>
                        <p className="font-medium">
                          {config.general.tick_time}s
                        </p>
                      </div>
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Flag TTL
                        </Label>
                        <p className="font-medium">
                          {config.general.flag_ttl} rounds
                        </p>
                      </div>
                    </div>
                    <div className="space-y-3">
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Start Time
                        </Label>
                        <p className="font-medium">
                          {new Date(config.general.start_time).toLocaleString()}
                        </p>
                      </div>
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          End Time
                        </Label>
                        <p className="font-medium">
                          {new Date(config.general.end_time).toLocaleString()}
                        </p>
                      </div>
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Total Ticks
                        </Label>
                        <p className="text-primary font-medium">
                          {calculateTotalTicks(config)}
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="flagChecker">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Flag Checker Configuration</CardTitle>
                  <CardDescription>
                    Settings for flag submission and validation
                  </CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => startEditing('flagChecker')}
                  disabled={editingSection === 'flagChecker'}
                >
                  <Edit3 className="mr-2 h-4 w-4" />
                  Edit
                </Button>
              </CardHeader>
              <CardContent className="space-y-4">
                {editingSection === 'flagChecker' && tempConfig ? (
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label>Flag Checker URL</Label>
                      <Input
                        value={tempConfig.flagChecker.url_flag_checker}
                        onChange={e =>
                          updateTempConfig(
                            'flagChecker',
                            'url_flag_checker',
                            e.target.value,
                          )
                        }
                      />
                      {errors['flagChecker.url_flag_checker'] && (
                        <p className="text-destructive text-xs">
                          {errors['flagChecker.url_flag_checker']}
                        </p>
                      )}
                    </div>
                    <div className="space-y-2">
                      <Label>Team Token</Label>
                      <Input
                        type="password"
                        value={tempConfig.flagChecker.team_token}
                        onChange={e =>
                          updateTempConfig(
                            'flagChecker',
                            'team_token',
                            e.target.value,
                          )
                        }
                      />
                      {errors['flagChecker.team_token'] && (
                        <p className="text-destructive text-xs">
                          {errors['flagChecker.team_token']}
                        </p>
                      )}
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label>Submit Time (seconds)</Label>
                        <Input
                          type="number"
                          value={
                            tempConfig.flagChecker.submit_flag_checker_time ||
                            ''
                          }
                          onChange={e =>
                            updateTempConfig(
                              'flagChecker',
                              'submit_flag_checker_time',
                              e.target.value,
                            )
                          }
                        />
                        {errors['flagChecker.submit_flag_checker_time'] && (
                          <p className="text-destructive text-xs">
                            {errors['flagChecker.submit_flag_checker_time']}
                          </p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label>Max Batch Size</Label>
                        <Input
                          type="number"
                          value={
                            tempConfig.flagChecker.max_flag_batch_size || ''
                          }
                          onChange={e =>
                            updateTempConfig(
                              'flagChecker',
                              'max_flag_batch_size',
                              e.target.value,
                            )
                          }
                        />
                        {errors['flagChecker.max_flag_batch_size'] && (
                          <p className="text-destructive text-xs">
                            {errors['flagChecker.max_flag_batch_size']}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="flex gap-2 pt-4">
                      <Button onClick={saveSection} disabled={isSaving}>
                        <Save className="mr-2 h-4 w-4" />
                        {isSaving ? 'Saving...' : 'Save Changes'}
                      </Button>
                      <Button variant="outline" onClick={cancelEditing}>
                        Cancel
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div>
                      <Label className="text-muted-foreground text-sm">
                        Flag Checker URL
                      </Label>
                      <p className="font-medium">
                        {config.flagChecker.url_flag_checker}
                      </p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground text-sm">
                        Team Token
                      </Label>
                      <p className="font-medium">
                        {'*'.repeat(config.flagChecker.team_token.length)}
                      </p>
                    </div>
                    <div className="grid grid-cols-2 gap-6">
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Submit Time
                        </Label>
                        <p className="font-medium">
                          {config.flagChecker.submit_flag_checker_time}s
                        </p>
                      </div>
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Max Batch Size
                        </Label>
                        <p className="font-medium">
                          {config.flagChecker.max_flag_batch_size}
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="flagInfo">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Flag Information</CardTitle>
                  <CardDescription>
                    Flag format and retrieval settings
                  </CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => startEditing('flagInfo')}
                  disabled={editingSection === 'flagInfo'}
                >
                  <Edit3 className="mr-2 h-4 w-4" />
                  Edit
                </Button>
              </CardHeader>
              <CardContent className="space-y-4">
                {editingSection === 'flagInfo' && tempConfig ? (
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label>Flag Regex Pattern</Label>
                      <Input
                        value={tempConfig.flagInfo.regex_flag}
                        onChange={e =>
                          updateTempConfig(
                            'flagInfo',
                            'regex_flag',
                            e.target.value,
                          )
                        }
                      />
                      {errors['flagInfo.regex_flag'] && (
                        <p className="text-destructive text-xs">
                          {errors['flagInfo.regex_flag']}
                        </p>
                      )}
                    </div>
                    <div className="space-y-2">
                      <Label>Flag IDs URL</Label>
                      <Input
                        value={tempConfig.flagInfo.url_flag_ids}
                        onChange={e =>
                          updateTempConfig(
                            'flagInfo',
                            'url_flag_ids',
                            e.target.value,
                          )
                        }
                      />
                      {errors['flagInfo.url_flag_ids'] && (
                        <p className="text-destructive text-xs">
                          {errors['flagInfo.url_flag_ids']}
                        </p>
                      )}
                    </div>
                    <div className="flex gap-2 pt-4">
                      <Button onClick={saveSection} disabled={isSaving}>
                        <Save className="mr-2 h-4 w-4" />
                        {isSaving ? 'Saving...' : 'Save Changes'}
                      </Button>
                      <Button variant="outline" onClick={cancelEditing}>
                        Cancel
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div>
                      <Label className="text-muted-foreground text-sm">
                        Flag Regex Pattern
                      </Label>
                      <p className="font-mono font-medium">
                        {config.flagInfo.regex_flag}
                      </p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground text-sm">
                        Flag IDs URL
                      </Label>
                      <p className="font-medium">
                        {config.flagInfo.url_flag_ids}
                      </p>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="services">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Services Configuration</CardTitle>
                  <CardDescription>
                    Manage monitored services and their ports
                  </CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => startEditing('services')}
                  disabled={editingSection === 'services'}
                >
                  <Edit3 className="mr-2 h-4 w-4" />
                  Edit
                </Button>
              </CardHeader>
              <CardContent className="space-y-4">
                {editingSection === 'services' && tempConfig ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <h4 className="font-medium">
                        Services ({tempConfig.services.length})
                      </h4>
                      <Button size="sm" onClick={addService}>
                        <Plus className="mr-2 h-4 w-4" />
                        Add Service
                      </Button>
                    </div>
                    {errors['services'] && (
                      <p className="text-destructive text-sm">
                        {errors['services']}
                      </p>
                    )}
                    <div className="space-y-3">
                      {tempConfig.services.map(service => (
                        <div
                          key={service.id}
                          className="flex items-center gap-3 rounded-lg border p-3"
                        >
                          <div className="grid flex-1 grid-cols-2 gap-3">
                            <div className="space-y-1">
                              <Label className="text-xs">Service Name</Label>
                              <Input
                                value={service.name}
                                onChange={e =>
                                  updateService(
                                    service.id,
                                    'name',
                                    e.target.value,
                                  )
                                }
                                placeholder="web"
                              />
                              {errors[`services.${service.id}.name`] && (
                                <p className="text-destructive text-xs">
                                  {errors[`services.${service.id}.name`]}
                                </p>
                              )}
                            </div>
                            <div className="space-y-1">
                              <Label className="text-xs">Port</Label>
                              <Input
                                type="number"
                                value={service.port || ''}
                                onChange={e =>
                                  updateService(
                                    service.id,
                                    'port',
                                    e.target.value,
                                  )
                                }
                                placeholder="80"
                              />
                              {errors[`services.${service.id}.port`] && (
                                <p className="text-destructive text-xs">
                                  {errors[`services.${service.id}.port`]}
                                </p>
                              )}
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
                    <div className="flex gap-2 pt-4">
                      <Button onClick={saveSection} disabled={isSaving}>
                        <Save className="mr-2 h-4 w-4" />
                        {isSaving ? 'Saving...' : 'Save Changes'}
                      </Button>
                      <Button variant="outline" onClick={cancelEditing}>
                        Cancel
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <h4 className="font-medium">
                        Services ({config.services.length})
                      </h4>
                    </div>
                    <div className="grid gap-3">
                      {config.services.map(service => (
                        <div
                          key={service.id}
                          className="flex items-center justify-between rounded-lg border p-3"
                        >
                          <div>
                            <p className="font-medium">{service.name}</p>
                            <p className="text-muted-foreground text-sm">
                              Port {service.port}
                            </p>
                          </div>
                          <Badge variant="secondary">
                            {service.name}:{service.port}
                          </Badge>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="teams">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Teams Configuration</CardTitle>
                  <CardDescription>
                    Team network and identification settings
                  </CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => startEditing('teams')}
                  disabled={editingSection === 'teams'}
                >
                  <Edit3 className="mr-2 h-4 w-4" />
                  Edit
                </Button>
              </CardHeader>
              <CardContent className="space-y-4">
                {editingSection === 'teams' && tempConfig ? (
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label>Number of Teams</Label>
                        <Input
                          type="number"
                          value={tempConfig.teams.range_ip_teams || ''}
                          onChange={e =>
                            updateTempConfig(
                              'teams',
                              'range_ip_teams',
                              e.target.value,
                            )
                          }
                        />
                        {errors['teams.range_ip_teams'] && (
                          <p className="text-destructive text-xs">
                            {errors['teams.range_ip_teams']}
                          </p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label>NOP Team ID</Label>
                        <Input
                          type="number"
                          value={tempConfig.teams.nop_team || ''}
                          onChange={e =>
                            updateTempConfig(
                              'teams',
                              'nop_team',
                              e.target.value,
                            )
                          }
                        />
                        {errors['teams.nop_team'] && (
                          <p className="text-destructive text-xs">
                            {errors['teams.nop_team']}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label>Your Team ID</Label>
                        <Input
                          type="number"
                          value={tempConfig.teams.my_team_id || ''}
                          onChange={e =>
                            updateTempConfig(
                              'teams',
                              'my_team_id',
                              e.target.value,
                            )
                          }
                        />
                        {errors['teams.my_team_id'] && (
                          <p className="text-destructive text-xs">
                            {errors['teams.my_team_id']}
                          </p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label>IP Format</Label>
                        <Input
                          value={tempConfig.teams.format_ip_teams}
                          onChange={e =>
                            updateTempConfig(
                              'teams',
                              'format_ip_teams',
                              e.target.value,
                            )
                          }
                        />
                        {errors['teams.format_ip_teams'] && (
                          <p className="text-destructive text-xs">
                            {errors['teams.format_ip_teams']}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="bg-muted rounded-lg p-3">
                      <h5 className="mb-2 font-medium">Example Team IPs:</h5>
                      <div className="text-muted-foreground space-y-1 text-sm">
                        <p>
                          Team 1:{' '}
                          {tempConfig.teams.format_ip_teams.replace('{}', '1')}
                        </p>
                        <p>
                          Team 2:{' '}
                          {tempConfig.teams.format_ip_teams.replace('{}', '2')}
                        </p>
                        <p>
                          Team 3:{' '}
                          {tempConfig.teams.format_ip_teams.replace('{}', '3')}
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-2 pt-4">
                      <Button onClick={saveSection} disabled={isSaving}>
                        <Save className="mr-2 h-4 w-4" />
                        {isSaving ? 'Saving...' : 'Save Changes'}
                      </Button>
                      <Button variant="outline" onClick={cancelEditing}>
                        Cancel
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-6">
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Total Teams
                        </Label>
                        <p className="font-medium">
                          {config.teams.range_ip_teams}
                        </p>
                      </div>
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          NOP Team ID
                        </Label>
                        <p className="font-medium">{config.teams.nop_team}</p>
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-6">
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          Your Team ID
                        </Label>
                        <p className="text-primary font-medium">
                          {config.teams.my_team_id}
                        </p>
                      </div>
                      <div>
                        <Label className="text-muted-foreground text-sm">
                          IP Format
                        </Label>
                        <p className="font-mono font-medium">
                          {config.teams.format_ip_teams}
                        </p>
                      </div>
                    </div>
                    <div className="bg-muted rounded-lg p-3">
                      <h5 className="mb-2 font-medium">Example Team IPs:</h5>
                      <div className="text-muted-foreground space-y-1 text-sm">
                        <p>
                          Team 1:{' '}
                          {config.teams.format_ip_teams.replace('{}', '1')}
                        </p>
                        <p>
                          Team 2:{' '}
                          {config.teams.format_ip_teams.replace('{}', '2')}
                        </p>
                        <p>
                          Team 3:{' '}
                          {config.teams.format_ip_teams.replace('{}', '3')}
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
