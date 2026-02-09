export function getServiceEntries(tagify) {
  return tagify.value.map(entry => {
    const [name, portStr] = entry.value.split(':');
    return {
      name: name.trim(),
      port: parseInt(portStr, 10)
    };
  }).filter(e => e.name && !isNaN(e.port));
}


export function validateConfigForm(document, tagify) {
  const validators = {
    team_token: val => val.length > 0,
    url_flag_checker: val => val.length > 0,
    protocol: val => val.length > 0,
    start_time: val => /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/.test(val),
    end_time: val => /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/.test(val),
    max_flag_batch_size: val => parseInt(val) > 0,
    tick_time: val => parseInt(val) > 0,
    submit_flag_checker_time: val => parseInt(val) >= 0,
    flag_ttl: val => parseInt(val) > 0,
    regex_flag: val => {
      try {
        new RegExp(val);
        return true;
      } catch (e) {
        return false;
      }
    },
    format_ip_teams: val => /^((\d{1,3}|\{\})\.){3}(\d{1,3}|\{\})$/.test(val),
    my_team_id: val => parseInt(val) >= 0,
    url_flag_ids: val => val.length > 0,
    nop_team: val => parseInt(val) >= 0,
    range_ip_teams: val => parseInt(val) > 0,
  };

  const resultBox = document.getElementById("config-result");

  for (const [id, check] of Object.entries(validators)) {
    const input = document.getElementById(id);
    const value = input?.value.trim();
    if (!value || !check(value)) {
      input?.focus();
      resultBox.textContent = `Invalid or missing: ${id.replace(/_/g, ' ')}`;
      resultBox.classList.add("text-red-500");
      return false;
    }
  }

  const services = getServiceEntries(tagify);
  if (services.length === 0) {
    document.getElementById("services").focus();
    resultBox.textContent = "Please provide at least one valid service (name:port)";
    return false;
  }

  return true;
}

export function buildConfigFromDOM(document, tagify) {
  const get = id => document.getElementById(id)?.value.trim();

  return {
    configured: true,
    server: {
      url_flag_checker: get("url_flag_checker"),
      team_token: get("team_token"),
      protocol: get("protocol"),
      start_time: get("start_time"),
      end_time: get("end_time"),
      max_flag_batch_size: Number(get("max_flag_batch_size")),
      tick_time: Number(get("tick_time")),
      submit_flag_checker_time: Number(get("submit_flag_checker_time")),
      flag_ttl: Number(get("flag_ttl")),
    },
    client: {
      services: getServiceEntries(tagify),
      regex_flag: get("regex_flag"),
      format_ip_teams: get("format_ip_teams"),
      my_team_id: Number(get("my_team_id")),
      url_flag_ids: get("url_flag_ids"),
      nop_team: Number(get("nop_team")),
      range_ip_teams: Number(get("range_ip_teams")),
    },
  };
}

export async function openModal() {
  document.getElementById('dialog').classList.remove('hidden');
  document.getElementById('table-head').classList.remove('sticky');
  await FillModalFields();
}
