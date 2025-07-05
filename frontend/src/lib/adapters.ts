import { ConfigData, SharedConfig } from './types';

export function sharedConfigToConfigData(
  shared_config: SharedConfig,
): ConfigData {
  return {
    general: {
      protocol: shared_config.server.protocol,
      tick_time: shared_config.server.tick_time,
      flag_ttl: shared_config.server.flag_ttl,
      start_time: shared_config.server.start_time,
      end_time: shared_config.server.end_time,
    },
    flagChecker: {
      url_flag_checker: shared_config.server.url_flag_checker,
      team_token: shared_config.server.team_token,
      submit_flag_checker_time: shared_config.server.submit_flag_checker_time,
      max_flag_batch_size: shared_config.server.max_flag_batch_size,
    },
    flagInfo: {
      regex_flag: shared_config.client.regex_flag,
      url_flag_ids: shared_config.client.url_flag_ids,
    },
    services: shared_config.client.services,
    teams: {
      range_ip_teams: shared_config.client.range_ip_teams,
      nop_team: shared_config.client.nop_team,
      my_team_id: shared_config.client.my_team_id,
      format_ip_teams: shared_config.client.format_ip_teams,
    },
  };
}

export function configDataToSharedConfig(configData: ConfigData): SharedConfig {
  return {
    configured: true,
    server: {
      protocol: configData.general.protocol,
      tick_time: Number(configData.general.tick_time),
      flag_ttl: Number(configData.general.flag_ttl),
      start_time: configData.general.start_time,
      end_time: configData.general.end_time,
      url_flag_checker: configData.flagChecker.url_flag_checker,
      team_token: configData.flagChecker.team_token,
      submit_flag_checker_time: Number(
        configData.flagChecker.submit_flag_checker_time,
      ),
      max_flag_batch_size: Number(configData.flagChecker.max_flag_batch_size),
    },
    client: {
      regex_flag: configData.flagInfo.regex_flag,
      url_flag_ids: configData.flagInfo.url_flag_ids,
      services: configData.services,
      range_ip_teams: configData.teams.range_ip_teams,
      nop_team: configData.teams.nop_team,
      my_team_id: Number(configData.teams.my_team_id),
      format_ip_teams: configData.teams.format_ip_teams,
    },
  };
}
