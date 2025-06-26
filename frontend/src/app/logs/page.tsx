"use client"

import { useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Search, Filter } from "lucide-react"

// Mock flag log data
const flagLogs = [
  {
    id: "FL001",
    code: "FLAG{web_exploit_alpha_r6}",
    team: { name: "Team Alpha", id: "team1", ip: "10.0.1.100" },
    status: "Accepted",
    executionTime: "2024-01-15 14:32:15",
    exploitTime: "2024-01-15 14:31:45",
  },
  {
    id: "FL002",
    code: "FLAG{db_injection_beta_r6}",
    team: { name: "Team Beta", id: "team2", ip: "10.0.2.100" },
    status: "Pending",
    executionTime: "2024-01-15 14:31:58",
    exploitTime: "2024-01-15 14:31:20",
  },
  {
    id: "FL003",
    code: "FLAG{api_overflow_gamma_r6}",
    team: { name: "Team Gamma", id: "team3", ip: "10.0.3.100" },
    status: "Denied",
    executionTime: "2024-01-15 14:31:42",
    exploitTime: "2024-01-15 14:31:10",
  },
  {
    id: "FL004",
    code: "FLAG{file_traversal_delta_r6}",
    team: { name: "Team Delta", id: "team4", ip: "10.0.4.100" },
    status: "Accepted",
    executionTime: "2024-01-15 14:31:28",
    exploitTime: "2024-01-15 14:30:55",
  },
  {
    id: "FL005",
    code: "FLAG{crypto_weak_alpha_r5}",
    team: { name: "Team Alpha", id: "team1", ip: "10.0.1.100" },
    status: "Accepted",
    executionTime: "2024-01-15 14:25:33",
    exploitTime: "2024-01-15 14:25:10",
  },
  {
    id: "FL006",
    code: "FLAG{network_sniff_beta_r5}",
    team: { name: "Team Beta", id: "team2", ip: "10.0.2.100" },
    status: "Denied",
    executionTime: "2024-01-15 14:24:17",
    exploitTime: "2024-01-15 14:23:45",
  },
  {
    id: "FL007",
    code: "FLAG{buffer_overflow_gamma_r5}",
    team: { name: "Team Gamma", id: "team3", ip: "10.0.3.100" },
    status: "Accepted",
    executionTime: "2024-01-15 14:23:52",
    exploitTime: "2024-01-15 14:23:20",
  },
  {
    id: "FL008",
    code: "FLAG{sql_injection_delta_r5}",
    team: { name: "Team Delta", id: "team4", ip: "10.0.4.100" },
    status: "Pending",
    executionTime: "2024-01-15 14:23:08",
    exploitTime: "2024-01-15 14:22:35",
  },
]

const getStatusBadge = (status: string) => {
  switch (status) {
    case "Accepted":
      return <Badge className="bg-green-500/10 text-green-500 border-green-500/20">Accepted</Badge>
    case "Pending":
      return <Badge className="bg-yellow-500/10 text-yellow-500 border-yellow-500/20">Pending</Badge>
    case "Denied":
      return <Badge className="bg-red-500/10 text-red-500 border-red-500/20">Denied</Badge>
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

export default function FlagLogs() {
  const [searchTerm, setSearchTerm] = useState("")
  const [statusFilter, setStatusFilter] = useState("All")
  const [teamFilter, setTeamFilter] = useState("All")

  const filteredLogs = flagLogs.filter((log) => {
    const matchesSearch =
      log.code.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.team.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.id.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesStatus = statusFilter === "All" || log.status === statusFilter
    const matchesTeam = teamFilter === "All" || log.team.name === teamFilter

    return matchesSearch && matchesStatus && matchesTeam
  })

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Flag Logs</h1>
          <p className="text-muted-foreground">View and manage all submitted flags with detailed information</p>
        </div>
        <Badge variant="outline" className="text-sm">
          {filteredLogs.length} flags
        </Badge>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            Filters & Search
          </CardTitle>
          <CardDescription>Filter flags by status, team, or search by flag code</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Search</label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search flags, teams, or IDs..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Status</label>
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="All">All Statuses</SelectItem>
                  <SelectItem value="Accepted">Accepted</SelectItem>
                  <SelectItem value="Pending">Pending</SelectItem>
                  <SelectItem value="Denied">Denied</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Team</label>
              <Select value={teamFilter} onValueChange={setTeamFilter}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="All">All Teams</SelectItem>
                  <SelectItem value="Team Alpha">Team Alpha</SelectItem>
                  <SelectItem value="Team Beta">Team Beta</SelectItem>
                  <SelectItem value="Team Gamma">Team Gamma</SelectItem>
                  <SelectItem value="Team Delta">Team Delta</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Flag Logs Table */}
      <Card>
        <CardHeader>
          <CardTitle>Flag Submissions</CardTitle>
          <CardDescription>Complete log of all flag submissions and their status</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Flag ID</TableHead>
                  <TableHead>Flag Code</TableHead>
                  <TableHead>Team</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Execution Time</TableHead>
                  <TableHead>Exploit Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredLogs.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="font-mono text-sm">{log.id}</TableCell>
                    <TableCell className="font-mono text-sm max-w-xs truncate">{log.code}</TableCell>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="font-medium">{log.team.name}</div>
                        <div className="text-xs text-muted-foreground">
                          {log.team.id} • {log.team.ip}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{getStatusBadge(log.status)}</TableCell>
                    <TableCell className="font-mono text-sm">{log.executionTime}</TableCell>
                    <TableCell className="font-mono text-sm">{log.exploitTime}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
