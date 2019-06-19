window.addEventListener('load', function() {
  var webAuth = new auth0.WebAuth({
    domain: AUTH_DOMAIN,
    clientID: EXTERNAL_ID,
    redirectUri: CALLBACK_URL,
    responseType: 'token id_token',
    scope: 'openid',
    leeway: 60
  })

  webAuth.parseHash(function(err, authResult) {
    if (authResult && authResult.accessToken && authResult.idToken) {
      sendToken(authResult.accessToken)
    } else if (err) {
      sendError(err)
    }
  })
})

function sendToken(token) {
  const message = {
    type: 'DECENTRALAND_USER_TOKEN',
    token: token,
  }
  const defaultAction =  function () {
    document.location.href = REDIRECT_BASE_URL + '?user_token=' + token
  }
  sendMessage(message, defaultAction)
}

function sendError(err) {
  const message = {
    type: 'DECENTRALAND_ERROR',
    error: err,
  }
  const defaultAction =  function () {
    console.log(err)
  }
  sendMessage(message, defaultAction)
}

function sendMessage(message, defaultAction) {
  const origin = APP_DOMAIN
  if (window.self !== window.top) {
    // Is within an Iframe
    message.from = 'IFRAME'
    window.parent.postMessage(message, origin)
  } else if (window.opener && window.opener !== window) {
    // Is within a Popup
    message.from = 'POPUP'
    window.opener.postMessage(message, origin)
  } else {
    // Not an iframe nor a popup, run default action
    defaultAction()
  }
}
