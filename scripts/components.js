(() => {
  "use strict";

  document.addEventListener("DOMContentLoaded", () => {
    const users = new Map();
    const scene = document.querySelector("a-scene");
    console.log(scene);

    function getRandomColor() {
      var letters = "0123456789ABCDEF";
      var color = "#";
      for (var i = 0; i < 6; i++) {
        color += letters[Math.floor(Math.random() * 16)];
      }
      return color;
    }

    function createUser(name, position) {
      if (users.has(name)) {
        return;
      }
      const el = document.createElement("a-sphere");
      el.setAttribute("position", position);
      el.setAttribute("radius", 1.25);
      el.setAttribute("color", getRandomColor());

      el.dataset.name = name;
      users.set(name, el);

      scene.append(el);
      console.log(`createUser: ${el}`);
    }

    function removeUser(name) {
      const el = users.get(name);
      if (!el) {
        return;
      }
      el.remove();
      users.delete(name);
      console.log(`removeUser: ${name}`);
    }

    function moveUser(name, position) {
      const el = users.get(name);
      if (!el) {
        return;
      }
      el.setAttribute("position", position);
      console.log(`moveUser: ${name}, ${position}`);
    }

    function getUserPosition(name) {
      const el = users.get(name);
      return el?.object3D?.position;
    }

    createUser(0, { x: 1, y: 0, z: 1 });
    setInterval(() => {
      const r = Math.random();
      if (r < 0.01) {
        const newPosition = {
          x: 2 - 4 * Math.random(),
          y: 0,
          z: 2 - 4 * Math.random(),
        };
        createUser(Math.random(), newPosition);
      } else if (0.01 <= r && r < 0.02) {
        const keys = users.keys().toArray();
        const key = keys[Math.ceil(keys.length * Math.random())];
        if (key) {
          removeUser(key);
        }
      } else {
        const keys = users.keys().toArray();
        const key = keys[Math.ceil(keys.length * Math.random())];
        if (key) {
          const position = getUserPosition(key);
          const newPosition = {
            x: position.x + (0.5 - Math.random()),
            y: position.y,
            z: position.z + (0.5 - Math.random()),
          };
          moveUser(key, newPosition);
        }
      }
    }, 10);
  });
})();
