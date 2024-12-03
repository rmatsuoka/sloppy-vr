const clientIdKey = "clientId";

export function getClientId() {
  const fromStorage = sessionStorage.getItem(clientIdKey);
  if (fromStorage != null) {
    return fromStorage;
  }

  const uuid = crypto.randomUUID();
  sessionStorage.setItem(clientIdKey, uuid);
  return uuid;
}
