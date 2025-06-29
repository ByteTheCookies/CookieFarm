'use client';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

// Mock data for service monitoring
const serviceData = [
  { round: 1, webService: 85, dbService: 92, apiService: 78, fileService: 88 },
  { round: 2, webService: 82, dbService: 89, apiService: 75, fileService: 85 },
  { round: 3, webService: 79, dbService: 86, apiService: 72, fileService: 82 },
  { round: 4, webService: 76, dbService: 83, apiService: 69, fileService: 79 },
  { round: 5, webService: 73, dbService: 80, apiService: 66, fileService: 76 },
  { round: 6, webService: 70, dbService: 77, apiService: 63, fileService: 73 },
];

const teams = [
  {
    name: 'Team Alpha',
    id: 'team1',
    ip: '10.0.1.100',
    services: [
      { name: 'Web Service', stolen: 15, lost: 8, uptime: 92 },
      { name: 'Database Service', stolen: 12, lost: 5, uptime: 98 },
      { name: 'API Service', stolen: 18, lost: 12, uptime: 85 },
      { name: 'File Service', stolen: 9, lost: 6, uptime: 94 },
    ],
  },
  {
    name: 'Team Beta',
    id: 'team2',
    ip: '10.0.2.100',
    services: [
      { name: 'Web Service', stolen: 22, lost: 6, uptime: 88 },
      { name: 'Database Service', stolen: 8, lost: 4, uptime: 96 },
      { name: 'API Service', stolen: 25, lost: 15, uptime: 82 },
      { name: 'File Service', stolen: 11, lost: 7, uptime: 91 },
    ],
  },
  {
    name: 'Team Gamma',
    id: 'team3',
    ip: '10.0.3.100',
    services: [
      { name: 'Web Service', stolen: 10, lost: 12, uptime: 95 },
      { name: 'Database Service', stolen: 6, lost: 8, uptime: 99 },
      { name: 'API Service', stolen: 14, lost: 18, uptime: 87 },
      { name: 'File Service', stolen: 7, lost: 9, uptime: 93 },
    ],
  },
  {
    name: 'Team Delta',
    id: 'team4',
    ip: '10.0.4.100',
    services: [
      { name: 'Web Service', stolen: 28, lost: 4, uptime: 86 },
      { name: 'Database Service', stolen: 15, lost: 3, uptime: 97 },
      { name: 'API Service', stolen: 32, lost: 8, uptime: 84 },
      { name: 'File Service', stolen: 19, lost: 5, uptime: 89 },
    ],
  },
];

export default function ServiceMonitoring() {
  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold">
            Service Monitoring
          </h1>
          <p className="text-muted-foreground">
            Monitor service performance and flag statistics across all teams
          </p>
        </div>
        <Badge variant="outline" className="text-sm">
          Live Monitoring
        </Badge>
      </div>

      {/* Service Performance Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Service Uptime by Round</CardTitle>
          <CardDescription>
            Average service uptime percentage across all teams
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={serviceData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="round"
                  className="text-muted-foreground"
                  tick={{ fontSize: 12 }}
                />
                <YAxis
                  className="text-muted-foreground"
                  tick={{ fontSize: 12 }}
                  domain={[0, 100]}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'hsl(var(--card))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                  }}
                />
                <Legend />
                <Bar dataKey="webService" fill="#ef4444" name="Web Service" />
                <Bar
                  dataKey="dbService"
                  fill="#3b82f6"
                  name="Database Service"
                />
                <Bar dataKey="apiService" fill="#10b981" name="API Service" />
                <Bar dataKey="fileService" fill="#f59e0b" name="File Service" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Team Service Details */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {teams.map(team => (
          <Card key={team.id}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-lg">{team.name}</CardTitle>
                  <CardDescription>
                    ID: {team.id} • IP: {team.ip}
                  </CardDescription>
                </div>
                <Badge variant="secondary">
                  {team.services.reduce(
                    (acc, service) => acc + service.uptime,
                    0,
                  ) / team.services.length}
                  % avg
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {team.services.map((service, index) => (
                  <div key={index} className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium">{service.name}</span>
                      <span className="text-muted-foreground">
                        {service.uptime}% uptime
                      </span>
                    </div>
                    <Progress value={service.uptime} className="h-2" />
                    <div className="text-muted-foreground flex justify-between text-xs">
                      <span className="text-red-400">
                        Stolen: {service.stolen}
                      </span>
                      <span className="text-orange-400">
                        Lost: {service.lost}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
