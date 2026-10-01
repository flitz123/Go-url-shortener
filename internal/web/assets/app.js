const storageKey = "linkline.links.v1";
const form = document.querySelector("#shorten-form");
const destinationInput = document.querySelector("#destination");
const aliasInput = document.querySelector("#custom-alias");
const aliasPrefix = document.querySelector("#alias-prefix");
const formError = document.querySelector("#form-error");
const linksList = document.querySelector("#links-list");
const emptyState = document.querySelector("#empty-state");
const clearButton = document.querySelector("#clear-history");
const totalLinks = document.querySelector("#total-links");
const totalClicks = document.querySelector("#total-clicks");
const linkCount = document.querySelector("#link-count");
const toast = document.querySelector("#toast");
let toastTimer;
let links = readLinks();

function readLinks() {
  try {
    const stored = JSON.parse(localStorage.getItem(storageKey) || "[]");
    return Array.isArray(stored) ? stored.filter((item) => item && item.code && item.original) : [];
  } catch {
    return [];
  }
}

function saveLinks() {
  localStorage.setItem(storageKey, JSON.stringify(links));
}

function notify(message) {
  toast.textContent = message;
  toast.classList.add("visible");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toast.classList.remove("visible"), 2300);
}

function setError(message) {
  formError.textContent = message;
  formError.hidden = !message;
}

function makeButton(label, action) {
  const button = document.createElement("button");
  button.type = "button";
  button.textContent = label;
  button.addEventListener("click", action);
  return button;
}

function formatDestination(value) {
  try {
    const parsed = new URL(value);
    return `${parsed.hostname}${parsed.pathname === "/" ? "" : parsed.pathname}${parsed.search}`;
  } catch {
    return value;
  }
}

function makeRow(item, index) {
  const row = document.createElement("article");
  row.className = "link-row";
  row.style.animationDelay = `${Math.min(index * 35, 210)}ms`;

  const primary = document.createElement("div");
  primary.className = "link-primary";
  const shortLink = document.createElement("a");
  shortLink.className = "short-link";
  shortLink.href = item.short_url || `${location.origin}/${encodeURIComponent(item.code)}`;
  shortLink.target = "_blank";
  shortLink.rel = "noopener noreferrer";
  shortLink.textContent = shortLink.href;
  const original = document.createElement("p");
  original.className = "original-link";
  original.textContent = formatDestination(item.original);
  original.title = item.original;
  primary.append(shortLink, original);

  const clicks = document.createElement("div");
  clicks.className = "click-count";
  clicks.textContent = Number(item.clicks || 0).toLocaleString();
  const clickLabel = document.createElement("span");
  clickLabel.textContent = "VISITS";
  clicks.append(clickLabel);

  const actions = document.createElement("div");
  actions.className = "row-actions";
  actions.append(makeButton("Copy link", () => copyLink(shortLink.href)));
  const openLink = document.createElement("a");
  openLink.href = shortLink.href;
  openLink.target = "_blank";
  openLink.rel = "noopener noreferrer";
  openLink.textContent = "Open ↗";
  actions.append(openLink);
  actions.append(makeButton("Hide", () => removeLink(item.code)));

  row.append(primary, clicks, actions);
  return row;
}

function renderLinks() {
  linksList.replaceChildren();
  emptyState.hidden = links.length > 0;
  if (links.length === 0) linksList.append(emptyState);
  links.forEach((item, index) => linksList.append(makeRow(item, index)));
  const clickTotal = links.reduce((sum, item) => sum + Number(item.clicks || 0), 0);
  totalLinks.textContent = links.length.toLocaleString();
  totalClicks.textContent = clickTotal.toLocaleString();
  linkCount.textContent = links.length.toLocaleString();
  clearButton.hidden = links.length === 0;
}

async function copyLink(value) {
  try {
    await navigator.clipboard.writeText(value);
    notify("Short link copied to clipboard.");
  } catch {
    notify("Clipboard access is unavailable in this browser.");
  }
}

function removeLink(code) {
  links = links.filter((item) => item.code !== code);
  saveLinks();
  renderLinks();
}

async function refreshClicks() {
  const updates = await Promise.all(links.map(async (item) => {
    try {
      const response = await fetch(`/api/urls/${encodeURIComponent(item.code)}`);
      if (!response.ok) return item;
      const current = await response.json();
      return { ...item, clicks: current.clicks, short_url: current.short_url };
    } catch {
      return item;
    }
  }));
  links = updates;
  saveLinks();
  renderLinks();
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  setError("");
  const enteredDestination = destinationInput.value.trim();
  const destination = /^[a-z][a-z\d+.-]*:\/\//i.test(enteredDestination)
    ? enteredDestination
    : `https://${enteredDestination}`;
  let parsed;
  try {
    parsed = new URL(destination);
  } catch {
    setError("Add a complete link, including https:// or http://.");
    destinationInput.focus();
    return;
  }
  if (!/^https?:$/.test(parsed.protocol) || !parsed.hostname) {
    setError("Only HTTP and HTTPS links can be shortened.");
    destinationInput.focus();
    return;
  }

  const createButton = form.querySelector("button[type='submit']");
  createButton.disabled = true;
  createButton.firstElementChild.textContent = "Making your short link…";
  try {
    const response = await fetch("/api/urls", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url: destination, custom_alias: aliasInput.value.trim() }),
    });
    if (!response.ok) {
      const message = await response.text();
      throw new Error(message.trim() || "Could not create that short link.");
    }
    const result = await response.json();
    links = [
      { ...result, created_at: new Date().toISOString() },
      ...links.filter((item) => item.code !== result.code),
    ].slice(0, 50);
    saveLinks();
    renderLinks();
    form.reset();
    aliasPrefix.textContent = `${location.host}/`;
    notify("Your short link is ready.");
    await copyLink(result.short_url);
    destinationInput.focus();
  } catch (error) {
    setError(error.message || "Something went wrong. Please try again.");
  } finally {
    createButton.disabled = false;
    createButton.firstElementChild.textContent = "Shorten this link";
  }
});

aliasPrefix.textContent = `${location.host}/`;
clearButton.addEventListener("click", () => {
  links = [];
  saveLinks();
  renderLinks();
});
renderLinks();
refreshClicks();