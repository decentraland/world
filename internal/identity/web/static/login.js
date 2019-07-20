
window.addEventListener('load', function() {
  var webAuth = new auth0.WebAuth({
    domain: DOMAIN,
    clientID: 'iRGF5TR5DBngi8yifjDGuHzixa9Q9HA8',
    redirectUri: CALLBACK_URL,
    responseType: 'token id_token',
    scope: 'openid',
    leeway: 60
  });

  webAuth.authorize();
})
