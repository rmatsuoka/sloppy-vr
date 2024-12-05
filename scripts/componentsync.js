export function startComponentSync(sock, myClientId) {
  console.log("startCompoenentSync");
  sock.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.clientId === myClientId) {
      return;
    }
    updateOrCreateUser(msg);
  };
  // const users = new Map();
  // setInterval(() => {
  //   const r = Math.random();
  //   const randomKey = () => {
  //     if (users.size === 0) {
  //       return undefined;
  //     }
  //     const keys = users.keys().toArray();
  //     return keys[Math.floor(keys.length * Math.random())];
  //   };
  //   if (r < 0.1) {
  //     users.set(Math.random(), {
  //       x: 2 - 4 * Math.random(),
  //       y: 0,
  //       z: 2 - 4 * Math.random(),
  //     });
  //   } else if (r < 0.2) {
  //     const key = randomKey();
  //     if (users.delete(key)) {
  //       console.log(`delete ${key}`);
  //     }
  //   } else {
  //     for (const [key, user] of users) {
  //       if (key !== undefined) {
  //         const user = users.get(key);
  //         users.set(key, {
  //           x: user.x + (0.5 - Math.random()),
  //           y: user.y,
  //           z: user.z + (0.5 - Math.random()),
  //         });
  //       }
  //     }
  //   }
  //   for (let [key, user] of users) {
  //     updateOrCreateUser({
  //       clientId: key,
  //       name: key,
  //       position: user,
  //     });
  //   }
  // }, 100);
}

function createUserElement({ clientId, name, position }) {
  const scene = document.querySelector("a-scene");
  const el = document.createElement("a-sphere");
  el.dataset.name = name;
  el.dataset.clientId = clientId;
  el.setAttribute("position", position);
  el.setAttribute("color", getRandomColor());
  el.setAttribute("radius", 1.25);
  el.setAttribute("delete-inactive", "");
  scene.append(el);
  return el;
}

function getUserElement(clientId) {
  return document.querySelector(`[data-client-id="${clientId}"]`);
}

function updateOrCreateUser({ clientId, name, position }) {
  let el = getUserElement(clientId);
  if (!el) {
    el = createUserElement({ clientId, name, position });
  }

  el.dataset.lastUpdated = Date.now();
  el.setAttribute("position", position);
}

function getRandomColor() {
  var letters = "0123456789ABCDEF";
  var color = "#";
  for (var i = 0; i < 6; i++) {
    color += letters[Math.floor(Math.random() * 16)];
  }
  return color;
}
