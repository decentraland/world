
window.addEventListener('load', function() {

  var webAuth = new auth0.WebAuth({
    domain: DOMAIN,
    clientID: EXTERNAL_ID,
    responseType: 'token id_token',
    scope: 'openid',
    leeway: 60
  });

  webAuth.logout({
    returnTo: CALLBACK_URL
  });
})
