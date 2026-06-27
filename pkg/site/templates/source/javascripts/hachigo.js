function getNav() {
  const mainNavs = document.querySelectorAll('ul.main-navigation, ul[role=main-navigation]');
  mainNavs.forEach(mainNav => {
    const fieldset = document.createElement('fieldset');
    fieldset.className = 'mobile-nav';
    const select = document.createElement('select');
    select.innerHTML = '<option value="">Navigate&hellip;</option>';

    const addOption = function(link) {
      const option = document.createElement('option');
      option.value = link.href;
      option.innerHTML = '&raquo; ' + link.textContent;
      select.appendChild(option);
    };

    mainNav.querySelectorAll('a').forEach(addOption);
    document.querySelectorAll('ul.subscription a').forEach(addOption);

    select.addEventListener('change', (event) => {
      if (event.target.value) {
        window.location.href = event.target.value;
      }
    });

    fieldset.appendChild(select);
    mainNav.parentNode.insertBefore(fieldset, mainNav);
  });
}

function addSidebarToggler() {
  if (!document.body.classList.contains('sidebar-footer')) {
    const content = document.getElementById('content');
    if (content) {
      const toggler = document.createElement('span');
      toggler.className = 'toggle-sidebar';
      toggler.addEventListener('click', (e) => {
        e.preventDefault();
        document.body.classList.toggle('collapse-sidebar');
      });
      content.appendChild(toggler);
    }
  }

  const sections = document.querySelectorAll('aside.sidebar > section');
  if (sections.length > 1) {
    sections.forEach((section, index) => {
      if (sections.length >= 3 && index % 3 === 0) {
        section.classList.add('first');
      }
      const count = ((index + 1) % 2) ? 'odd' : 'even';
      section.classList.add(count);
    });
  }
  if (sections.length >= 3) {
    const sidebar = document.querySelector('aside.sidebar');
    if (sidebar) sidebar.classList.add('thirds');
  }
}

function testFeatures() {
  const html = document.documentElement;
  const style = document.createElement('div').style;
  const hasMaskImage = 'maskImage' in style || 
                       'webkitMaskImage' in style || 
                       'MozMaskImage' in style || 
                       'msMaskImage' in style || 
                       'OMaskImage' in style;
  if (hasMaskImage) {
    html.classList.add('maskImage');
  } else {
    html.classList.add('no-maskImage');
  }
  if ('placeholder' in document.createElement('input')) {
    html.classList.add('placeholder');
  } else {
    html.classList.add('no-placeholder');
  }
}

function addCodeLineNumbers() {
  if (navigator.appName === 'Microsoft Internet Explorer') { return; }
  document.querySelectorAll('div.gist-highlight').forEach(code => {
    const tableStart = '<table><tbody><tr><td class="gutter">';
    let lineNumbers = '<pre class="line-numbers">';
    const tableMiddle = '</pre></td><td class="code">';
    const tableEnd = '</td></tr></tbody></table>';
    const lines = code.querySelectorAll('.line');
    const count = lines.length;
    for (let i = 1; i <= count; i++) {
      lineNumbers += '<span class="line-number">' + i + '</span>\n';
    }
    const pre = code.querySelector('pre');
    if (pre) {
      const table = tableStart + lineNumbers + tableMiddle + '<pre>' + pre.innerHTML + '</pre>' + tableEnd;
      code.innerHTML = table;
    }
  });
}

function wrapFlashVideos() {
  document.querySelectorAll('object').forEach(obj => {
    if (obj.querySelector('param[name=movie]')) {
      const wrapper = document.createElement('div');
      wrapper.className = 'flash-video';
      obj.parentNode.insertBefore(wrapper, obj);
      wrapper.appendChild(obj);
    }
  });

  document.querySelectorAll('iframe[src*="vimeo"], iframe[src*="youtube"]').forEach(iframe => {
    const wrapper = document.createElement('div');
    wrapper.className = 'flash-video';
    iframe.parentNode.insertBefore(wrapper, iframe);
    wrapper.appendChild(iframe);
  });
}

function renderDeliciousLinks(items) {
  let output = "<ul>";
  for (let i = 0; i < items.length; i++) {
    output += '<li><a href="' + items[i].u + '" title="Tags: ' + (items[i].t == "" ? "" : items[i].t.join(', ')) + '">' + items[i].d + '</a></li>';
  }
  output += "</ul>";
  const delicious = document.getElementById('delicious');
  if (delicious) {
    delicious.innerHTML = output;
  }
}

document.addEventListener('DOMContentLoaded', () => {
  testFeatures();
  wrapFlashVideos();
  addCodeLineNumbers();
  getNav();
  addSidebarToggler();
});

// iOS scaling bug fix
(function(doc) {
  var addEvent = 'addEventListener',
      type = 'gesturestart',
      qsa = 'querySelectorAll',
      scales = [1, 1],
      meta = qsa in doc ? doc[qsa]('meta[name=viewport]') : [];
  function fix() {
    meta.content = 'width=device-width,minimum-scale=' + scales[0] + ',maximum-scale=' + scales[1];
    doc.removeEventListener(type, fix, true);
  }
  if ((meta = meta[meta.length - 1]) && addEvent in doc) {
    fix();
    scales = [0.25, 1.6];
    doc[addEvent](type, fix, true);
  }
}(document));
