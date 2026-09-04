// private-tab.uc.mjs — приватная вкладка вместо приватного окна.
//
// Перехватывает window.OpenBrowserWindow({private: true}) — единую точку входа,
// через которую Firefox/Zen открывают приватное окно (пункт меню, Ctrl+Shift+P,
// всё остальное стоковое). Вместо нового окна открывает вкладку в этом же окне,
// в СВОЁМ одноразовом контейнере (contextual identity) на каждую вкладку —
// в отличие от Waterfox, где все "приватные" вкладки делят один контейнер и,
// следовательно, куки друг друга. Контейнер и все его данные (куки, localStorage,
// IndexedDB, кэш) удаляются, когда вкладка закрывается — через штатный
// ContextualIdentityService.remove(), который сам чистит данные под капотом.
//
// Загружается через fx-autoconfig (см. builder/vendor/fx-autoconfig).

const { ContextualIdentityService } = ChromeUtils.importESModule(
  "resource://gre/modules/ContextualIdentityService.sys.mjs"
);
const { startupFinished } = ChromeUtils.importESModule(
  "chrome://userchromejs/content/utils.sys.mjs"
);

const IDENTITY_NAME = "Приватная вкладка";
const IDENTITY_ICON = "fingerprint";
const IDENTITY_COLOR = "purple";

// userContextId наших вкладок — чтобы не спутать с обычными контейнерами
// пользователя (Personal/Work/Banking и т.д.) при уборке.
const shadowContexts = new Set();

function openPrivateTab() {
  const identity = ContextualIdentityService.create(
    IDENTITY_NAME,
    IDENTITY_ICON,
    IDENTITY_COLOR
  );
  shadowContexts.add(identity.userContextId);

  const tab = gBrowser.addTab("about:blank", {
    userContextId: identity.userContextId,
    triggeringPrincipal: Services.scriptSecurityManager.getSystemPrincipal(),
  });
  gBrowser.selectedTab = tab;
  return tab;
}

function cleanupContextForTab(tab) {
  const userContextId = tab.userContextId;
  if (!userContextId || !shadowContexts.has(userContextId)) {
    return;
  }
  shadowContexts.delete(userContextId);
  try {
    // remove() сам вызывает Services.clearData.deleteDataFromOriginAttributesPattern
    // для этого userContextId — отдельно чистить куки/storage не нужно.
    ContextualIdentityService.remove(userContextId);
  } catch (ex) {
    console.error("[private-tab] не смог очистить контейнер:", ex);
  }
}

function purgeOrphanedContexts() {
  // Если браузер закрыли принудительно (сбой, force-quit) до TabClose,
  // контейнер от прошлой сессии остаётся с данными внутри. Распознаём
  // такие по имени/иконке/цвету — их не могло создать ничего, кроме нас —
  // и подчищаем при старте.
  for (const identity of ContextualIdentityService.getPublicIdentities()) {
    if (
      identity.name === IDENTITY_NAME &&
      identity.icon === IDENTITY_ICON &&
      identity.color === IDENTITY_COLOR
    ) {
      try {
        ContextualIdentityService.remove(identity.userContextId);
      } catch (ex) {
        console.error("[private-tab] не смог убрать осиротевший контейнер:", ex);
      }
    }
  }
}

async function init() {
  await startupFinished();
  purgeOrphanedContexts();

  const originalOpenBrowserWindow = window.OpenBrowserWindow;
  window.OpenBrowserWindow = function (options) {
    if (options && options.private) {
      return openPrivateTab();
    }
    return originalOpenBrowserWindow.apply(this, arguments);
  };

  gBrowser.tabContainer.addEventListener("TabClose", (event) => {
    cleanupContextForTab(event.target);
  });
}

init();
