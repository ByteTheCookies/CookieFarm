export async function checkConfig() {
  try {
    const data = await getConfig();
    if (data.error) {
      return false;
    }
    if (!data.configured) {
      return false;
    }
  } catch (error) {
    console.error(error);
    return false;
  }
  return true;
}

export async function sendConfig(config) {
  const res = await fetch("/api/v1/config", {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({
      config: config,
    }),
  });

  if (!res.ok) {
    return res.status;
  }
}

export async function getConfig() {
  try {
    const res = await fetch("/api/v1/config", {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });
    const data = await res.json();
    return data;
  } catch (error) {
    console.error(error);
    return { "error": error.message };
  }
}


export async function sendFlag(flag) {
  const res = await fetch("/api/v1/submit-flag", {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({
      flag: flag,
    }),
  });
  return res
}

export async function deleteFlag(flag) {
  const res = await fetch(`/api/v1/delete-flag?flag=${flag}`, {
    method: 'DELETE',
    credentials: 'include',
  });
  return res
}
