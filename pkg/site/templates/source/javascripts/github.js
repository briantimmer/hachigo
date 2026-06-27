var github = (function(){
  function escapeHtml(str) {
    var div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }
  function render(target, repos){
    var i = 0, fragment = '', t = document.querySelector(target);
    if (!t) return;

    for(i = 0; i < repos.length; i++) {
      fragment += '<li><a href="'+repos[i].html_url+'">'+repos[i].name+'</a><p>'+escapeHtml(repos[i].description||'')+'</p></li>';
    }
    t.innerHTML = fragment;
  }
  return {
    showRepos: function(options){
      fetch("https://api.github.com/users/"+options.user+"/repos?sort=pushed")
        .then(function(res) {
          if (!res.ok) throw new Error("HTTP error " + res.status);
          return res.json();
        })
        .then(function(data) {
          var repos = [];
          if (!data) { return; }
          for (var i = 0; i < data.length; i++) {
            if (options.skip_forks && data[i].fork) { continue; }
            repos.push(data[i]);
          }
          if (options.count) { repos.splice(options.count); }
          render(options.target, repos);
        })
        .catch(function(err) {
          var loadingLi = document.querySelector(options.target + ' li.loading');
          if (loadingLi) {
            loadingLi.classList.add('error');
            loadingLi.textContent = "Error loading feed";
          }
        });
    }
  };
})();
