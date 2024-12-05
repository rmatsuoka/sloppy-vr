import { aframeRegisterComponent } from "./components.js";
import { getClientId } from "./clientid.js";
import { startComponentSync } from "./componentsync.js";

const clientId = getClientId();
const sock = new WebSocket(`wss://${window.location.host}/socketserver`);

sock.onerror = (event) => {
  console.log(event);
};

aframeRegisterComponent(sock, clientId);

document.addEventListener("DOMContentLoaded", () => {
  startComponentSync(sock, clientId);
});
