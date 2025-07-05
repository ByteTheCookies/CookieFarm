'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import useSWR from 'swr';
import axios from 'axios';

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';

import {
  Info,
  Plus,
  Trash2,
  Check,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { z } from 'zod';
import { useRouter } from 'next/navigation';

import { Service, ConfigData, Protocol } from '@/lib/types'; // Adjust the import path as needed
import { fetcher } from '@/lib/fetchers';

const initialConfig: ConfigData = {
  general: {
    protocol: '',
    tick_time: 120,
    flag_ttl: 5,
    start_time: '2025-07-02T10:00',
    end_time: '2025-07-02T18:00',
  },
  flagChecker: {
    url_flag_checker: 'http://localhost:5001/flags',
    team_token: '1234567890abcdef1234567890abcdef',
    submit_flag_checker_time: 120,
    max_flag_batch_size: 1000,
  },
  flagInfo: {
    regex_flag: '[A-Z0-9]{31}=',
    url_flag_ids: 'http://localhost:5001/flagIds',
  },
  services: [],
  teams: {
    range_ip_teams: 29,
    nop_team: 0,
    my_team_id: 1,
    format_ip_teams: '10.10.{}.1',
  },
};

const steps = [
  {
    id: 1,
    title: 'General Settings',
  },
  {
    id: 2,
    title: 'Flag Checker Setup',
  },
  {
    id: 3,
    title: 'Flag Information',
  },
  {
    id: 4,
    title: 'Services',
  },
  {
    id: 5,
    title: 'Teams Info',
  },
];

// Zod schemas for validation
const serviceSchema = z.object({
  id: z.string(),
  name: z.string().min(1, 'Service name is required'),
  port: z.number().min(1, 'Port must be greater than 0'),
});

const generalSchema = z.object({
  protocol: z.string().min(1, 'Protocol is required'),
  tick_time: z.number().min(1, 'Tick time must be greater than 0'),
  flag_ttl: z.number().min(0, 'Flag time to live must positive'),
  start_time: z.string().min(1, 'Start time is required'),
  end_time: z.string().min(1, 'End time is required'),
});

const flagCheckerSchema = z.object({
  url_flag_checker: z
    .string()
    .min(1, 'url_flag_checker is required')
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
    .min(1, 'Flag IDs address is required')
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

const configSchema = z.object({
  general: generalSchema,
  flagChecker: flagCheckerSchema,
  flagInfo: flagInfoSchema,
  services: z.array(serviceSchema).min(1, 'At least one service is required'),
  teams: teamsSchema,
});

export function ServerConfigStepper() {
  const [currentStep, setCurrentStep] = useState(1);
  const [config, setConfig] = useState<ConfigData>(initialConfig);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [globalError, setGlobalError] = useState<string | null>(null);
  const [completed, setCompleted] = useState(false);
  const router = useRouter();
  const {
    data: protocols = [],
    error: protocolsError,
    isLoading: loadingProtocols,
  } = useSWR<Protocol[]>(`/api/protocols`, fetcher);

  const updateConfig = (
    section: keyof ConfigData,
    field: string,
    value: any, //eslint-disable-line @typescript-eslint/no-explicit-any
  ) => {
    let processedValue = value;

    // Convert string numbers to actual numbers for numeric fields
    if (section === 'general' && ['tick_time', 'flag_ttl'].includes(field)) {
      processedValue = value === '' ? 0 : parseInt(value) || 0;
    }
    if (
      section === 'flagChecker' &&
      ['submit_flag_checker_time', 'max_flag_batch_size'].includes(field)
    ) {
      processedValue = value === '' ? 0 : parseInt(value) || 0;
    }
    if (
      section === 'teams' &&
      ['range_ip_teams', 'nop_team', 'my_team_id'].includes(field)
    ) {
      processedValue = value === '' ? 0 : parseInt(value) || 0;
    }

    setConfig(prev => ({
      ...prev,
      [section]: {
        ...prev[section],
        [field]: processedValue,
      },
    }));
    setErrors(prev => {
      const newErrors = { ...prev };
      delete newErrors[`${section}.${field}`];
      return newErrors;
    });
    setGlobalError(null);
  };

  const addService = () => {
    const newService: Service = {
      id: Date.now().toString(),
      name: '',
      port: 0,
    };
    setConfig(prev => ({
      ...prev,
      services: [...prev.services, newService],
    }));
    setGlobalError(null);
  };

  const removeService = (id: string) => {
    setConfig(prev => ({
      ...prev,
      services: prev.services.filter(service => service.id !== id),
    }));
    setGlobalError(null);
  };

  const updateService = (
    id: string,
    field: keyof Service,
    value: string | number,
  ) => {
    let processedValue = value;

    // Convert string to number for port field
    if (field === 'port') {
      processedValue =
        typeof value === 'string'
          ? value === ''
            ? 0
            : parseInt(value) || 0
          : value;
    }

    setConfig(prev => ({
      ...prev,
      services: prev.services.map(service =>
        service.id === id ? { ...service, [field]: processedValue } : service,
      ),
    }));
    setErrors(prev => {
      const newErrors = { ...prev };
      delete newErrors[`services.${id}.${field}`];
      return newErrors;
    });
    setGlobalError(null);
  };

  const calculateTotalTicks = () => {
    const { start_time, end_time, tick_time } = config.general;
    if (!start_time || !end_time || !tick_time) return 0;

    const start = new Date(start_time).getTime();
    const end = new Date(end_time).getTime();
    const tickMs = tick_time * 1000;

    if (end <= start || tickMs <= 0) return 0;

    return Math.floor((end - start) / tickMs);
  };

  // Validation per step
  const validateStep = (): boolean => {
    const stepErrors: Record<string, string> = {};
    setGlobalError(null);

    if (currentStep === 1) {
      const result = generalSchema.safeParse(config.general);
      if (!result.success) {
        for (const issue of result.error.issues) {
          stepErrors[`general.${issue.path[0]}`] = issue.message;
        }
      }
    } else if (currentStep === 2) {
      const result = flagCheckerSchema.safeParse(config.flagChecker);
      if (!result.success) {
        for (const issue of result.error.issues) {
          stepErrors[`flagChecker.${issue.path[0]}`] = issue.message;
        }
      }
    } else if (currentStep === 3) {
      const result = flagInfoSchema.safeParse(config.flagInfo);
      if (!result.success) {
        for (const issue of result.error.issues) {
          stepErrors[`flagInfo.${issue.path[0]}`] = issue.message;
        }
      }
    } else if (currentStep === 4) {
      if (config.services.length === 0) {
        setGlobalError('At least one service is required.');
        return false;
      }
      config.services.forEach(service => {
        const result = serviceSchema.safeParse(service);
        if (!result.success) {
          for (const issue of result.error.issues) {
            stepErrors[`services.${service.id}.${issue.path[0]}`] =
              issue.message;
          }
        }
      });
    } else if (currentStep === 5) {
      const result = teamsSchema.safeParse(config.teams);
      if (!result.success) {
        for (const issue of result.error.issues) {
          stepErrors[`teams.${issue.path[0]}`] = issue.message;
        }
      }
    }
    setErrors(stepErrors);
    setGlobalError(Object.values(stepErrors)[0] || null);
    return Object.keys(stepErrors).length === 0;
  };

  const nextStep = () => {
    if (validateStep()) {
      if (currentStep < steps.length) {
        setCurrentStep(currentStep + 1);
      }
    }
  };

  const prevStep = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
      setGlobalError(null);
    }
  };

  // Handle completion
  const handleComplete = () => {
    // Validate all steps
    const result = configSchema.safeParse(config);
    if (!result.success) {
      // Find first error and go to that step
      const firstIssue = result.error.issues[0];
      if (firstIssue.path[0] === 'general') setCurrentStep(1);
      else if (firstIssue.path[0] === 'flagChecker') setCurrentStep(2);
      else if (firstIssue.path[0] === 'flagInfo') setCurrentStep(3);
      else if (firstIssue.path[0] === 'services') setCurrentStep(4);
      else if (firstIssue.path[0] === 'teams') setCurrentStep(5);
      // Set errors for all
      const allErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        if (
          issue.path[0] === 'services' &&
          typeof issue.path[1] === 'number' &&
          issue.path[2]
        ) {
          // Service array error
          const service = config.services[issue.path[1]];
          allErrors[`services.${service?.id}.${issue.path[2]}`] = issue.message;
        } else {
          allErrors[`${issue.path[0]}.${issue.path[1] ?? ''}`] = issue.message;
        }
      }
      setErrors(allErrors);
      setGlobalError(
        result.error.issues[0]?.message || 'Please fix errors above.',
      );
      setCompleted(false);
      return;
    }
    setErrors({});
    setGlobalError(null);
    setCompleted(true);

    const configFormatted = {
      config: {
        server: {
          ...config.general,
          ...config.flagChecker,
        },
        client: {
          services: config.services.map(({ id, ...service }) => service), //eslint-disable-line
          ...config.teams,
          ...config.flagInfo,
        },
        configured: true,
      },
    };

    // Send as JSON instead of FormData
    axios
      .post('/api/config', configFormatted, {
        headers: {
          'Content-Type': 'application/json',
        },
      })
      .then(response => {
        console.log('Config submitted successfully:', response.data);
        // Puoi aggiungere qui altre azioni dopo il successo dell'invio
        router.push('/');
      })
      .catch(error => {
        console.error('Config submission failed:', error);
        // Puoi gestire l'errore qui, ad esempio mostrando un messaggio all'utente
        setGlobalError('Failed to submit configuration. Please try again.');
        setCompleted(false);
      });
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

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <div className="space-y-6">
            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="protocol">Communication Protocol</Label>
                <InfoTooltip content="The protocol used to communicate with the flag checker (depends on CTF platform)" />
              </div>
              <Select
                value={config.general.protocol}
                onValueChange={value =>
                  updateConfig('general', 'protocol', value)
                }
                disabled={loadingProtocols}
              >
                <SelectTrigger>
                  <SelectValue
                    placeholder={
                      loadingProtocols
                        ? 'Loading protocols...'
                        : protocolsError
                          ? 'Error loading protocols'
                          : 'Select protocol'
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {protocols.map(protocol => (
                    <SelectItem key={protocol.value} value={protocol.value}>
                      {protocol.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {protocolsError && (
                <p className="text-destructive text-xs">
                  Error loading protocols: {protocolsError.message}
                </p>
              )}
              {errors['general.protocol'] && (
                <p className="text-destructive text-xs">
                  {errors['general.protocol']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="tick-time">Tick Time (seconds)</Label>
                <InfoTooltip content="Duration of each round/tick in seconds. Common values are 60, 120, or 300 seconds" />
              </div>
              <Input
                id="tick-time"
                placeholder="120"
                type="number"
                value={config.general.tick_time || ''}
                onChange={e =>
                  updateConfig('general', 'tick_time', e.target.value)
                }
              />
              {errors['general.tick_time'] && (
                <p className="text-destructive text-xs">
                  {errors['general.tick_time']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="flag-ttl">Flag Time to Live (rounds)</Label>
                <InfoTooltip content="How long flags remain valid for submission after being generated" />
              </div>
              <Input
                id="flag-ttl"
                placeholder="5"
                type="number"
                value={config.general.flag_ttl || 0}
                onChange={e =>
                  updateConfig('general', 'flag_ttl', e.target.value)
                }
              />
              {errors['general.flag_ttl'] && (
                <p className="text-destructive text-xs">
                  {errors['general.flag_ttl']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="start-time">Match Start Time</Label>
                <InfoTooltip content="When the CTF match begins. Use local time format" />
              </div>
              <Input
                id="start-time"
                type="datetime-local"
                value={config.general.start_time}
                onChange={e =>
                  updateConfig('general', 'start_time', e.target.value)
                }
              />
              {errors['general.start_time'] && (
                <p className="text-destructive text-xs">
                  {errors['general.start_time']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="end-time">Match End Time</Label>
                <InfoTooltip content="When the CTF match ends. Use local time format" />
              </div>
              <Input
                id="end-time"
                type="datetime-local"
                value={config.general.end_time}
                onChange={e =>
                  updateConfig('general', 'end_time', e.target.value)
                }
              />
              {errors['general.end_time'] && (
                <p className="text-destructive text-xs">
                  {errors['general.end_time']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label>Total Number of Ticks</Label>
                <InfoTooltip content="Automatically calculated based on match duration and tick time" />
              </div>
              <div className="bg-muted rounded-md p-3">
                <span className="text-primary text-2xl font-bold">
                  {calculateTotalTicks()}
                </span>
                <span className="text-muted-foreground ml-2 text-sm">
                  ticks
                </span>
              </div>
            </div>
            {globalError && (
              <p className="text-destructive text-sm">{globalError}</p>
            )}
          </div>
        );

      case 2:
        return (
          <div className="space-y-6">
            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="url_flag_checker">url_flag_checker</Label>
                <InfoTooltip content="The address of the flag checker server where flags will be submitted, in the format ip:port (e.g., 192.168.1.100:8080)" />
              </div>
              <Input
                id="url_flag_checker"
                placeholder="http://192.168.1.100:8080"
                value={config.flagChecker.url_flag_checker}
                onChange={e =>
                  updateConfig(
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
              {errors['flagChecker.team_token'] && (
                <p className="text-destructive text-xs">
                  {errors['flagChecker.team_token']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="submit-time">Submit Time (seconds)</Label>
                <InfoTooltip content="How many seconds to wait before submitting all pending flags in batch" />
              </div>
              <Input
                id="submit-time"
                placeholder="30"
                type="number"
                value={config.flagChecker.submit_flag_checker_time || ''}
                onChange={e =>
                  updateConfig(
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
              <div className="flex items-center">
                <Label htmlFor="batch-size">Batch Size</Label>
                <InfoTooltip content="Maximum number of flags to submit in a single batch request" />
              </div>
              <Input
                id="batch-size"
                placeholder="100"
                type="number"
                value={config.flagChecker.max_flag_batch_size || ''}
                onChange={e =>
                  updateConfig(
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
            {globalError && (
              <p className="text-destructive text-sm">{globalError}</p>
            )}
          </div>
        );

      case 3:
        return (
          <div className="space-y-6">
            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="flag-format">Flag Format</Label>
                <InfoTooltip content="The regex pattern or format used to identify flags (e.g., FLAG{...} or CTF{...})" />
              </div>
              <Input
                id="flag-format"
                placeholder="FLAG{.*}"
                value={config.flagInfo.regex_flag}
                onChange={e =>
                  updateConfig('flagInfo', 'regex_flag', e.target.value)
                }
              />
              {errors['flagInfo.regex_flag'] && (
                <p className="text-destructive text-xs">
                  {errors['flagInfo.regex_flag']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="flag-ids-url_flag_checker">
                  Flag IDs url_flag_checker
                </Label>
                <InfoTooltip content="URL or endpoint where the flag IDs can be retrieved for each tick" />
              </div>
              <Input
                id="flag-ids-url_flag_checker"
                placeholder="https://192.168.1.100:8080/flagids"
                type="url"
                value={config.flagInfo.url_flag_ids}
                onChange={e =>
                  updateConfig('flagInfo', 'url_flag_ids', e.target.value)
                }
              />
              {errors['flagInfo.url_flag_ids'] && (
                <p className="text-destructive text-xs">
                  {errors['flagInfo.url_flag_ids']}
                </p>
              )}
            </div>

            <div className="bg-muted mt-6 rounded-lg p-4">
              <h4 className="mb-2 font-semibold">Flag Information Summary:</h4>
              <div className="text-muted-foreground space-y-1 text-sm">
                <p>
                  <strong>Format:</strong>{' '}
                  {config.flagInfo.regex_flag || 'Not set'}
                </p>
                <p>
                  <strong>IDs Source:</strong>{' '}
                  {config.flagInfo.url_flag_ids || 'Not set'}
                </p>
              </div>
            </div>
            {globalError && (
              <p className="text-destructive text-sm">{globalError}</p>
            )}
          </div>
        );

      case 4:
        return (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold">
                  Services Configuration
                </h3>
                <p className="text-muted-foreground text-sm">
                  Add the services that will be monitored during the CTF
                </p>
              </div>
              <Button onClick={addService} size="sm">
                <Plus className="mr-2 h-4 w-4" />
                Add Service
              </Button>
            </div>

            {config.services.length === 0 ? (
              <div className="text-muted-foreground py-8 text-center">
                <p>No services configured yet.</p>
                <p className="text-sm">
                  Click &quot;Add Service&quot; to get started.
                </p>
              </div>
            ) : (
              <div className="mx-auto w-6/12 space-y-4">
                {config.services.map(service => (
                  <Card key={service.id} className="py-4">
                    <CardContent>
                      <div className="flex items-center gap-4">
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
                                updateService(
                                  service.id,
                                  'name',
                                  e.target.value,
                                )
                              }
                            />
                            {errors[`services.${service.id}.name`] && (
                              <p className="text-destructive text-xs">
                                {errors[`services.${service.id}.name`]}
                              </p>
                            )}
                          </div>
                          <div className="space-y-2">
                            <div className="flex items-center">
                              <Label>Service Port</Label>
                              <InfoTooltip content="Port number on which this service runs on team machines" />
                            </div>
                            <Input
                              placeholder="80"
                              type="number"
                              value={service.port || ''}
                              onChange={e =>
                                updateService(
                                  service.id,
                                  'port',
                                  e.target.value,
                                )
                              }
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
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
            {globalError && (
              <p className="text-destructive text-sm">{globalError}</p>
            )}
          </div>
        );

      case 5:
        return (
          <div className="space-y-6">
            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="num-teams">Number of Teams</Label>
                <InfoTooltip content="Total number of teams participating in the CTF" />
              </div>
              <Input
                id="num-teams"
                placeholder="10"
                type="number"
                value={config.teams.range_ip_teams || ''}
                onChange={e =>
                  updateConfig('teams', 'range_ip_teams', e.target.value)
                }
              />
              {errors['teams.range_ip_teams'] && (
                <p className="text-destructive text-xs">
                  {errors['teams.range_ip_teams']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="nop-team">NOP Team ID</Label>
                <InfoTooltip content="ID of the NOP (No Operation) team - usually used for testing or as a baseline" />
              </div>
              <Input
                id="nop-team"
                placeholder="0"
                type="number"
                value={config.teams.nop_team || 0}
                onChange={e =>
                  updateConfig('teams', 'nop_team', e.target.value)
                }
              />
              {errors['teams.nop_team'] && (
                <p className="text-destructive text-xs">
                  {errors['teams.nop_team']}
                </p>
              )}
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
                value={config.teams.my_team_id || ''}
                onChange={e =>
                  updateConfig('teams', 'my_team_id', e.target.value)
                }
              />
              {errors['teams.my_team_id'] && (
                <p className="text-destructive text-xs">
                  {errors['teams.my_team_id']}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="ip-format">Team IP Format</Label>
                <InfoTooltip content="IP address pattern for teams. Use {} as placeholder for team ID (e.g., 10.10.{}.1 becomes 10.10.5.1 for team 5)" />
              </div>
              <Input
                id="ip-format"
                placeholder="10.10.{}.1"
                value={config.teams.format_ip_teams}
                onChange={e =>
                  updateConfig('teams', 'format_ip_teams', e.target.value)
                }
              />
              {errors['teams.format_ip_teams'] && (
                <p className="text-destructive text-xs">
                  {errors['teams.format_ip_teams']}
                </p>
              )}
            </div>

            <div className="bg-muted mt-6 rounded-lg p-4">
              <h4 className="mb-2 font-semibold">Example Team IPs:</h4>
              <div className="text-muted-foreground space-y-1 text-sm">
                <p>Team 1: {config.teams.format_ip_teams.replace('{}', '1')}</p>
                <p>Team 2: {config.teams.format_ip_teams.replace('{}', '2')}</p>
                <p>Team 3: {config.teams.format_ip_teams.replace('{}', '3')}</p>
              </div>
            </div>
            {globalError && (
              <p className="text-destructive text-sm">{globalError}</p>
            )}
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="space-y-8">
      {/* Stepper Header */}
      <div className="flex items-center justify-between">
        {steps.map((step, index) => (
          <div key={step.id} className="flex items-center">
            <div className="flex items-center">
              <div
                className={`flex h-10 w-10 items-center justify-center rounded-full border-2 ${
                  currentStep === step.id
                    ? 'border-primary bg-primary text-primary-foreground'
                    : currentStep > step.id
                      ? 'border-green-500 bg-green-500 text-white'
                      : 'border-muted-foreground bg-background text-muted-foreground'
                }`}
              >
                {currentStep > step.id ? (
                  <Check className="h-5 w-5" />
                ) : (
                  <span className="text-sm font-semibold">{step.id}</span>
                )}
              </div>
              <div className="ml-3 hidden sm:block">
                <p
                  className={`text-sm font-medium ${
                    currentStep === step.id
                      ? 'text-primary'
                      : 'text-muted-foreground'
                  }`}
                >
                  {step.title}
                </p>
              </div>
            </div>
            {index < steps.length - 1 && (
              <div className="bg-border mx-4 hidden h-px flex-1 sm:block" />
            )}
          </div>
        ))}
      </div>

      {/* Step Content */}
      <Card>
        <CardHeader>
          <CardTitle>{steps[currentStep - 1].title}</CardTitle>
        </CardHeader>
        <CardContent>{renderStepContent()}</CardContent>
      </Card>

      {/* Navigation */}
      <div className="flex items-center justify-between">
        <Button
          variant="outline"
          onClick={prevStep}
          disabled={currentStep === 1}
        >
          <ChevronLeft className="mr-2 h-4 w-4" />
          Previous
        </Button>

        <div className="flex items-center gap-2">
          <span className="text-muted-foreground text-sm">
            Step {currentStep} of {steps.length}
          </span>
        </div>

        {currentStep === steps.length ? (
          <Button
            className="bg-green-600 hover:bg-green-700"
            onClick={handleComplete}
          >
            <Check className="mr-2 h-4 w-4" />
            Complete Setup
          </Button>
        ) : (
          <Button onClick={nextStep}>
            Next
            <ChevronRight className="ml-2 h-4 w-4" />
          </Button>
        )}
      </div>

      {/* Configuration Summary (shown on last step) */}
      {currentStep === steps.length && (
        <Card>
          <CardHeader>
            <CardTitle>Configuration Summary</CardTitle>
            <CardDescription>
              Review your server configuration before completing setup
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <h4 className="mb-2 font-semibold">General Settings</h4>
              <div className="text-muted-foreground space-y-1 text-sm">
                <p>Protocol: {config.general.protocol}</p>
                <p>Tick Time: {config.general.tick_time}s</p>
                <p>Flag Time to Live: {config.general.flag_ttl}s</p>
                <p>Total Ticks: {calculateTotalTicks()}</p>
              </div>
            </div>
            <Separator />
            <div>
              <h4 className="mb-2 font-semibold">Flag Checker</h4>
              <div className="text-muted-foreground space-y-1 text-sm">
                <p>Server: {config.flagChecker.url_flag_checker}</p>
                <p>
                  Submit Time: {config.flagChecker.submit_flag_checker_time}s,
                  Batch Size: {config.flagChecker.max_flag_batch_size}
                </p>
              </div>
            </div>
            <Separator />
            <div>
              <h4 className="mb-2 font-semibold">Flag Information</h4>
              <div className="text-muted-foreground space-y-1 text-sm">
                <p>Format: {config.flagInfo.regex_flag}</p>
                <p>FlagIDs Address: {config.flagInfo.url_flag_ids}</p>
              </div>
            </div>
            <Separator />
            <div>
              <h4 className="mb-2 font-semibold">
                Services ({config.services.length})
              </h4>
              <div className="text-muted-foreground text-sm">
                {config.services.map(service => (
                  <p key={service.id}>
                    {service.name}:{service.port}
                  </p>
                ))}
              </div>
            </div>
            <Separator />
            <div>
              <h4 className="mb-2 font-semibold">Teams</h4>
              <div className="text-muted-foreground space-y-1 text-sm">
                <p>Total Teams: {config.teams.range_ip_teams}</p>
                <p>
                  Your Team: {config.teams.my_team_id}, NOP Team:{' '}
                  {config.teams.nop_team}
                </p>
                <p>IP Format: {config.teams.format_ip_teams}</p>
              </div>
            </div>
            {completed && (
              <div className="mt-4 rounded bg-green-100 p-3 text-green-800">
                Configuration complete! Check the console for the output.
              </div>
            )}
            {globalError && (
              <div className="mt-4 rounded bg-red-100 p-3 text-red-800">
                {globalError}
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
