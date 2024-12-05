async function getMy() {
  const res = await fetch("/my");
  if (res.status == 404) {
    return null;
  }
  if (300 <= res.status && res.status < 600) {
    throw new Error(await res.text());
  }
  return await res.json();
}

async function main() {
  const div = document.querySelector("#signin-or-entervr");
  if (div === null) {
    throw new Error("cannot find div.signin-or-eventvr");
  }

  const user = await getMy();
  if (user === null) {
    div.innerHTML = `<a href="/signin">Sign in</a>`;
  } else {
    div.innerHTML = `<p>ようこそ ${user.display_name}</p>
    <a href="/vr">enter metaverse</a>`;
  }
}

document.addEventListener("DOMContentLoaded", () => {
  main().catch((x) => console.log(x));
});
