import { Button } from "@cloudflare/kumo/components/button";
import { Input } from "@cloudflare/kumo/components/input";
import { Select } from "@cloudflare/kumo/components/select";
import { Tabs } from "@cloudflare/kumo/components/tabs";
import type { ConfigServices } from "@/api/config";

const statusTabs = [
  { value: "all", label: "All" },
  { value: "0", label: "Queued" },
  { value: "1", label: "Accepted" },
  { value: "2", label: "Denied" },
  { value: "3", label: "Error" },
  { value: "4", label: "Invalid" },
] as const;

export type FlagFilterState = {
  status: (typeof statusTabs)[number]["value"];
  service: string;
  team: string;
  search: string;
  searchField: string;
};

export function FlagFilters(props: Readonly<{
  filters: FlagFilterState;
  services: ConfigServices;
  onChange: (nextFilters: FlagFilterState) => void;
}>) {
  return (
    <section className="space-y-4 rounded-2xl border border-kumo-line bg-kumo-base p-4">
      <div className="flex items-center justify-between">
        <Tabs
          variant="segmented"
          tabs={statusTabs.map((tab) => ({
            value: tab.value,
            label: tab.label,
          }))}
          value={props.filters.status}
          onValueChange={(value) => {
            props.onChange({
              ...props.filters,
              status: value as FlagFilterState["status"],
            });
          }}
        />
        <Button
          variant="secondary"
          onClick={() => {
            props.onChange({
              status: "all",
              service: "",
              team: "",
              search: "",
              searchField: "flag_code",
            });
          }}
        >
          Reset
        </Button>
      </div>

      <div className="grid gap-4 xl:grid-cols-4">
        <Select
          label="Service"
          placeholder="All services"
          value={props.filters.service || "all"}
          renderValue={(value) => (value === "all" ? "All services" : String(value))}
          onValueChange={(value) => {
            props.onChange({
              ...props.filters,
              service: String(value) === "all" ? "" : String(value),
            });
          }}
        >
          <Select.Option value="all">All services</Select.Option>
          {Object.keys(props.services)
            .sort((left, right) => left.localeCompare(right))
            .map((serviceName) => (
              <Select.Option key={serviceName} value={serviceName}>
                {serviceName}
              </Select.Option>
            ))}
        </Select>

        <Input
          label="Team ID"
          placeholder="1"
          value={props.filters.team}
          onChange={(event) => {
            props.onChange({
              ...props.filters,
              team: event.target.value,
            });
          }}
        />

        <Input
          label="Search"
          placeholder="Flag, message, username..."
          value={props.filters.search}
          onChange={(event) => {
            props.onChange({
              ...props.filters,
              search: event.target.value,
            });
          }}
        />

        <Select
          label="Search field"
          value={props.filters.searchField}
          onValueChange={(value) => {
            props.onChange({
              ...props.filters,
              searchField: String(value),
            });
          }}
        >
          <Select.Option value="flag_code">Flag</Select.Option>
          <Select.Option value="msg">Message</Select.Option>
          <Select.Option value="service_name">Service</Select.Option>
          <Select.Option value="username">Username</Select.Option>
          <Select.Option value="exploit_name">Exploit</Select.Option>
        </Select>
      </div>
    </section>
  );
}
