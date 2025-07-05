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
