// This type is used to define the shape of our data.
// You can use a Zod schema here if you want.
export type Flag = {
  flag_code: string;
  service_name: string;
  status: 'ACCEPTED' | 'DENIED' | 'RESUBMIT' | 'ERROR' | 'UNSUBMITTED';
  exploit_name: string;
  msg: string;
  submit_time: number;
  response_time: number;
  port_service: number;
  team_id: number;
};

export type Protocol = {
  value: string;
  label: string;
};

export type Service = {
  id: string;
  name: string;
  port: number;
};

export type SharedConfig = {
  server: {
    protocol: Protocol;
    tick_time: number;
    flag_ttl: number;
    start_time: string;
    end_time: string;
    url_flag_checker: string;
    team_token: string;
    submit_flag_checker_time: number;
    max_flag_batch_size: number;
  };
  client: {
    regex_flag: string;
    url_flag_ids: string;
    range_ip_teams: number;
    nop_team: number;
    my_team_id: number;
    format_ip_teams: string;
    services: Service[];
  };
};

export type ConfigData = {
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
};
