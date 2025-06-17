"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Info, Plus, Trash2, Check, ChevronLeft, ChevronRight } from "lucide-react"

interface Service {
  id: string
  name: string
  port: string
}

interface ConfigData {
  flagChecker: {
    ipAddress: string
    port: string
    teamToken: string
    submitTime: string
    batchSize: string
    protocol: string
  }
  rounds: {
    duration: string
    startTime: string
    endTime: string
  }
  services: Service[]
  teams: {
    numberOfTeams: string
    nopTeamId: string
    ownTeamId: string
    ipFormat: string
  }
}

const initialConfig: ConfigData = {
  flagChecker: {
    ipAddress: "",
    port: "",
    teamToken: "",
    submitTime: "30",
    batchSize: "100",
    protocol: "",
  },
  rounds: {
    duration: "120",
    startTime: "",
    endTime: "",
  },
  services: [],
  teams: {
    numberOfTeams: "",
    nopTeamId: "",
    ownTeamId: "",
    ipFormat: "10.10.{}.1",
  },
}

const protocols = [
  { value: "cc_http", label: "CC HTTP" },
  { value: "cc_grpc", label: "CC gRPC" },
  { value: "defcon_http", label: "DEFCON HTTP" },
  { value: "custom_tcp", label: "Custom TCP" },
  { value: "custom_udp", label: "Custom UDP" },
]

const steps = [
  { id: 1, title: "Flag Checker Setup", description: "Configure flag submission settings" },
  { id: 2, title: "Rounds Info", description: "Set up match timing and rounds" },
  { id: 3, title: "Services", description: "Define services to monitor" },
  { id: 4, title: "Teams Info", description: "Configure team settings" },
]

export function ServerConfigStepper() {
  const [currentStep, setCurrentStep] = useState(1)
  const [config, setConfig] = useState<ConfigData>(initialConfig)

  const updateConfig = (section: keyof ConfigData, field: string, value: any) => { //eslint-disable-line @typescript-eslint/no-explicit-any
    setConfig((prev) => ({
      ...prev,
      [section]: {
        ...prev[section],
        [field]: value,
      },
    }))
  }

  const addService = () => {
    const newService: Service = {
      id: Date.now().toString(),
      name: "",
      port: "",
    }
    setConfig((prev) => ({
      ...prev,
      services: [...prev.services, newService],
    }))
  }

  const removeService = (id: string) => {
    setConfig((prev) => ({
      ...prev,
      services: prev.services.filter((service) => service.id !== id),
    }))
  }

  const updateService = (id: string, field: keyof Service, value: string) => {
    setConfig((prev) => ({
      ...prev,
      services: prev.services.map((service) => (service.id === id ? { ...service, [field]: value } : service)),
    }))
  }

  const calculateTotalTicks = () => {
    const { startTime, endTime, duration } = config.rounds
    if (!startTime || !endTime || !duration) return 0

    const start = new Date(startTime).getTime()
    const end = new Date(endTime).getTime()
    const durationMs = Number.parseInt(duration) * 1000

    if (end <= start || durationMs <= 0) return 0

    return Math.floor((end - start) / durationMs)
  }

  const nextStep = () => {
    if (currentStep < steps.length) {
      setCurrentStep(currentStep + 1)
    }
  }

  const prevStep = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1)
    }
  }

  const InfoTooltip = ({ content }: { content: string }) => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Info className="h-4 w-4 text-muted-foreground cursor-help ml-2" />
        </TooltipTrigger>
        <TooltipContent className="max-w-xs">
          <p>{content}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <div className="space-y-6">
            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="ip-address">IP Address</Label>
                <InfoTooltip content="The IP address of the flag checker server where flags will be submitted" />
              </div>
              <Input
                id="ip-address"
                placeholder="192.168.1.100"
                value={config.flagChecker.ipAddress}
                onChange={(e) => updateConfig("flagChecker", "ipAddress", e.target.value)}
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
                onChange={(e) => updateConfig("flagChecker", "port", e.target.value)}
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
                value={config.flagChecker.teamToken}
                onChange={(e) => updateConfig("flagChecker", "teamToken", e.target.value)}
              />
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
                value={config.flagChecker.submitTime}
                onChange={(e) => updateConfig("flagChecker", "submitTime", e.target.value)}
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
                onChange={(e) => updateConfig("flagChecker", "batchSize", e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="protocol">Communication Protocol</Label>
                <InfoTooltip content="The protocol used to communicate with the flag checker (depends on CTF platform)" />
              </div>
              <Select
                value={config.flagChecker.protocol}
                onValueChange={(value) => updateConfig("flagChecker", "protocol", value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select protocol" />
                </SelectTrigger>
                <SelectContent>
                  {protocols.map((protocol) => (
                    <SelectItem key={protocol.value} value={protocol.value}>
                      {protocol.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        )

      case 2:
        return (
          <div className="space-y-6">
            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="round-duration">Round Duration (seconds)</Label>
                <InfoTooltip content="Duration of each round/tick in seconds. Common values are 60, 120, or 300 seconds" />
              </div>
              <Input
                id="round-duration"
                placeholder="120"
                type="number"
                value={config.rounds.duration}
                onChange={(e) => updateConfig("rounds", "duration", e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label htmlFor="start-time">Match Start Time</Label>
                <InfoTooltip content="When the CTF match begins. Use local time format" />
              </div>
              <Input
                id="start-time"
                type="datetime-local"
                value={config.rounds.startTime}
                onChange={(e) => updateConfig("rounds", "startTime", e.target.value)}
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
                onChange={(e) => updateConfig("rounds", "endTime", e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <div className="flex items-center">
                <Label>Total Number of Ticks</Label>
                <InfoTooltip content="Automatically calculated based on match duration and round length" />
              </div>
              <div className="p-3 bg-muted rounded-md">
                <span className="text-2xl font-bold text-primary">{calculateTotalTicks()}</span>
                <span className="text-sm text-muted-foreground ml-2">ticks</span>
              </div>
            </div>
          </div>
        )

      case 3:
        return (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold">Services Configuration</h3>
                <p className="text-sm text-muted-foreground">Add the services that will be monitored during the CTF</p>
              </div>
              <Button onClick={addService} size="sm">
                <Plus className="h-4 w-4 mr-2" />
                Add Service
              </Button>
            </div>

            {config.services.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <p>No services configured yet.</p>
                <p className="text-sm">Click &quot; Add Service&quot; to get started.</p>
              </div>
            ) : (
              <div className="space-y-4">
                {config.services.map((service, index) => (
                  <Card key={service.id}>
                    <CardContent className="pt-6">
                      <div className="flex items-center gap-4">
                        <Badge variant="outline">Service {index + 1}</Badge>
                        <div className="flex-1 grid grid-cols-1 md:grid-cols-2 gap-4">
                          <div className="space-y-2">
                            <div className="flex items-center">
                              <Label>Service Name</Label>
                              <InfoTooltip content="Name of the service (e.g., web, database, api)" />
                            </div>
                            <Input
                              placeholder="web"
                              value={service.name}
                              onChange={(e) => updateService(service.id, "name", e.target.value)}
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
                              onChange={(e) => updateService(service.id, "port", e.target.value)}
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
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </div>
        )

      case 4:
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
                value={config.teams.numberOfTeams}
                onChange={(e) => updateConfig("teams", "numberOfTeams", e.target.value)}
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
                value={config.teams.nopTeamId}
                onChange={(e) => updateConfig("teams", "nopTeamId", e.target.value)}
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
                onChange={(e) => updateConfig("teams", "ownTeamId", e.target.value)}
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
                onChange={(e) => updateConfig("teams", "ipFormat", e.target.value)}
              />
            </div>

            <div className="mt-6 p-4 bg-muted rounded-lg">
              <h4 className="font-semibold mb-2">Example Team IPs:</h4>
              <div className="space-y-1 text-sm text-muted-foreground">
                <p>Team 1: {config.teams.ipFormat.replace("{}", "1")}</p>
                <p>Team 2: {config.teams.ipFormat.replace("{}", "2")}</p>
                <p>Team 3: {config.teams.ipFormat.replace("{}", "3")}</p>
              </div>
            </div>
          </div>
        )

      default:
        return null
    }
  }

  return (
    <div className="space-y-8">
      {/* Stepper Header */}
      <div className="flex items-center justify-between">
        {steps.map((step, index) => (
          <div key={step.id} className="flex items-center">
            <div className="flex items-center">
              <div
                className={`flex items-center justify-center w-10 h-10 rounded-full border-2 ${currentStep === step.id
                  ? "border-primary bg-primary text-primary-foreground"
                  : currentStep > step.id
                    ? "border-green-500 bg-green-500 text-white"
                    : "border-muted-foreground bg-background text-muted-foreground"
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
                  className={`text-sm font-medium ${currentStep === step.id ? "text-primary" : "text-muted-foreground"
                    }`}
                >
                  {step.title}
                </p>
                <p className="text-xs text-muted-foreground">{step.description}</p>
              </div>
            </div>
            {index < steps.length - 1 && <div className="flex-1 mx-4 h-px bg-border hidden sm:block" />}
          </div>
        ))}
      </div>

      {/* Step Content */}
      <Card>
        <CardHeader>
          <CardTitle>{steps[currentStep - 1].title}</CardTitle>
          <CardDescription>{steps[currentStep - 1].description}</CardDescription>
        </CardHeader>
        <CardContent>{renderStepContent()}</CardContent>
      </Card>

      {/* Navigation */}
      <div className="flex items-center justify-between">
        <Button variant="outline" onClick={prevStep} disabled={currentStep === 1}>
          <ChevronLeft className="h-4 w-4 mr-2" />
          Previous
        </Button>

        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">
            Step {currentStep} of {steps.length}
          </span>
        </div>

        {currentStep === steps.length ? (
          <Button className="bg-green-600 hover:bg-green-700">
            <Check className="h-4 w-4 mr-2" />
            Complete Setup
          </Button>
        ) : (
          <Button onClick={nextStep}>
            Next
            <ChevronRight className="h-4 w-4 ml-2" />
          </Button>
        )}
      </div>

      {/* Configuration Summary (shown on last step) */}
      {currentStep === steps.length && (
        <Card>
          <CardHeader>
            <CardTitle>Configuration Summary</CardTitle>
            <CardDescription>Review your server configuration before completing setup</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <h4 className="font-semibold mb-2">Flag Checker</h4>
              <div className="text-sm text-muted-foreground space-y-1">
                <p>
                  Server: {config.flagChecker.ipAddress}:{config.flagChecker.port}
                </p>
                <p>Protocol: {config.flagChecker.protocol}</p>
                <p>
                  Submit Time: {config.flagChecker.submitTime}s, Batch Size: {config.flagChecker.batchSize}
                </p>
              </div>
            </div>
            <Separator />
            <div>
              <h4 className="font-semibold mb-2">Match Info</h4>
              <div className="text-sm text-muted-foreground space-y-1">
                <p>Round Duration: {config.rounds.duration}s</p>
                <p>Total Ticks: {calculateTotalTicks()}</p>
              </div>
            </div>
            <Separator />
            <div>
              <h4 className="font-semibold mb-2">Services ({config.services.length})</h4>
              <div className="text-sm text-muted-foreground">
                {config.services.map((service, index) => ( // eslint-disable-line
                  <p key={service.id}>
                    {service.name}:{service.port}
                  </p>
                ))}
              </div>
            </div>
            <Separator />
            <div>
              <h4 className="font-semibold mb-2">Teams</h4>
              <div className="text-sm text-muted-foreground space-y-1">
                <p>Total Teams: {config.teams.numberOfTeams}</p>
                <p>
                  Your Team: {config.teams.ownTeamId}, NOP Team: {config.teams.nopTeamId}
                </p>
                <p>IP Format: {config.teams.ipFormat}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
