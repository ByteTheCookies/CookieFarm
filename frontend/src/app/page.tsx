'use client';

import { useState } from 'react';
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
import { Badge } from '@/components/ui/badge';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

// Mock data for the chart
const mockData = [
  { round: 1, team1: 12, team2: 8, team3: 15, team4: 6 },
  { round: 2, team1: 18, team2: 14, team3: 22, team4: 10 },
  { round: 3, team1: 25, team2: 19, team3: 28, team4: 16 },
  { round: 4, team1: 32, team2: 26, team3: 35, team4: 22 },
  { round: 5, team1: 38, team2: 31, team3: 42, team4: 28 },
  { round: 6, team1: 45, team2: 37, team3: 48, team4: 34 },
];

const teams = [
  'All Teams',
  'Team Alpha',
  'Team Beta',
  'Team Gamma',
  'Team Delta',
];
const exploits = [
  'All Exploits',
  'Web Exploit',
  'Binary Exploit',
  'Crypto Exploit',
  'Network Exploit',
];
const services = [
  'All Services',
  'Web Service',
  'Database Service',
  'API Service',
  'File Service',
];

export default function MatchOverview() {
  const [selectedTeam, setSelectedTeam] = useState('All Teams');
  const [selectedExploit, setSelectedExploit] = useState('All Exploits');
  const [selectedService, setSelectedService] = useState('All Services');

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold">Match Overview</h1>
          <p className="text-muted-foreground">
            Monitor flag capture progress across all teams and rounds
          </p>
        </div>
        <Badge variant="outline" className="text-sm">
          Round 6 Active
        </Badge>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
          <CardDescription>
            Filter the data by team, exploit type, and service
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <label className="text-sm font-medium">Team</label>
              <Select value={selectedTeam} onValueChange={setSelectedTeam}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {teams.map(team => (
                    <SelectItem key={team} value={team}>
                      {team}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Exploit</label>
              <Select
                value={selectedExploit}
                onValueChange={setSelectedExploit}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {exploits.map(exploit => (
                    <SelectItem key={exploit} value={exploit}>
                      {exploit}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Service</label>
              <Select
                value={selectedService}
                onValueChange={setSelectedService}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {services.map(service => (
                    <SelectItem key={service} value={service}>
                      {service}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Flags Captured by Round</CardTitle>
          <CardDescription>
            Number of flags captured by each team across rounds
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={mockData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="round"
                  className="text-muted-foreground"
                  tick={{ fontSize: 12 }}
                />
                <YAxis
                  className="text-muted-foreground"
                  tick={{ fontSize: 12 }}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'hsl(var(--card))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                  }}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="team1"
                  stroke="#ef4444"
                  strokeWidth={2}
                  name="Team Alpha"
                />
                <Line
                  type="monotone"
                  dataKey="team2"
                  stroke="#3b82f6"
                  strokeWidth={2}
                  name="Team Beta"
                />
                <Line
                  type="monotone"
                  dataKey="team3"
                  stroke="#10b981"
                  strokeWidth={2}
                  name="Team Gamma"
                />
                <Line
                  type="monotone"
                  dataKey="team4"
                  stroke="#f59e0b"
                  strokeWidth={2}
                  name="Team Delta"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Flags</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">1,247</div>
            <p className="text-muted-foreground text-xs">
              +12% from last round
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Active Teams</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">4</div>
            <p className="text-muted-foreground text-xs">All teams online</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Current Round</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">6</div>
            <p className="text-muted-foreground text-xs">2:34 remaining</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">78%</div>
            <p className="text-muted-foreground text-xs">+5% from average</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
