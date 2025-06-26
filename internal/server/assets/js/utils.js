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
    host_flagchecker: val => val.length > 0,
    protocol: val => val.length > 0,
    my_team_ip: val => /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/.test(val),
    regex_flag: val => {
      try {
        new RegExp(val);
        return true;
      } catch (e) {
        return false;
      }
    },
    format_ip_teams: val => /^((\d{1,3}|\{\})\.){3}(\d{1,3}|\{\})$/.test(val),
    range_ip_teams: val => parseInt(val) > 0,
    max_flag_batch_size: val => parseInt(val) > 0,
    submit_flag_checker_time: val => parseInt(val) >= 0,
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
      team_token: get("team_token"),
      host_flagchecker: get("host_flagchecker"),
      protocol: get("protocol"),
      max_flag_batch_size: Number(get("max_flag_batch_size")),
      submit_flag_checker_time: Number(get("submit_flag_checker_time")),
    },
    client: {
      services: getServiceEntries(tagify),
      range_ip_teams: Number(get("range_ip_teams")),
      format_ip_teams: get("format_ip_teams"),
      my_team_ip: get("my_team_ip"),
      regex_flag: get("regex_flag"),
    },
  };
}

export async function openModal() {
  document.getElementById('dialog').classList.remove('hidden');
  document.getElementById('table-head').classList.remove('sticky');
  await FillModalFields();
}
