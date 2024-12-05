import * as socktype from "./socktype.js";

export function aframeRegisterComponent(sock, clientId) {
  console.log("aframe");
  AFRAME.registerComponent("send-position", {
    init: function () {
      this.timestamp = Date.now();
    },

    tick: function () {
      if (Date.now() - this.timestamp > 17) {
        this.timestamp = Date.now();
        const buf = JSON.stringify({
          type: socktype.Position,
          clientId: clientId,
          name: "",
          position: this.el.object3D.position,
        });
        sock.send(buf);
        // console.log(buf);
      }
    },
  });

  AFRAME.registerComponent("delete-inactive", {
    tick: function () {
      if (inactiveUserElement(this.el)) {
        this.el.remove();
        // console.log(`this.el.remove: ${this.el.dataset.clientId}`);
      }
    },
  });
}

function inactiveUserElement(userElement) {
  const lastUpdated = Number(userElement.dataset.lastUpdated);
  const deadline = Date.now() - 5 * 1000;
  return lastUpdated < deadline;
}
