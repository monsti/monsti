(function() {
  var monstiFileChooser = function(field_name, url, type, win) {
    var cmsURL = window.location.toString();    // script URL - use an absolute path!
    cmsURL = cmsURL.replace("@@edit", "@@chooser?type=" + type);
//    alert("Field_Name: " + field_name + " URL: " + url + " Type: " + type + " Win: " + win + " CMSURL: " + cmsURL); // debug/testing

    tinyMCE.activeEditor.windowManager.open({
      file : cmsURL,
      title : 'File Chooser',
      width : 700,
      height : 500,
      resizable : "yes",
      inline : "yes",  // This parameter only has an effect if you use the inlinepopups plugin!
      close_previous : "no"
    }, {
      window : win,
      input : field_name
    });
    return false;
  }

  $(document).ready(function () {
    tinymce.init({
      selector: ".html-field textarea",
      plugins: "anchor autosave code hr image visualchars visualblocks table paste media link",
      tools: "inserttable",
      height: 300,
      file_browser_callback: monstiFileChooser,
      language : monsti.session.locale,
      formats : {
        alignleft : {selector : 'p,h1,h2,h3,h4,h5,h6,td,th,div,ul,ol,li,table,img', classes : 'monsti--htmlarea-align-left'},
        aligncenter : {selector : 'p,h1,h2,h3,h4,h5,h6,td,th,div,ul,ol,li,table,img', classes : 'monsti--htmlarea-align-center'},
        alignright : {selector : 'p,h1,h2,h3,h4,h5,h6,td,th,div,ul,ol,li,table,img', classes : 'monsti--htmlarea-align-right'},
        alignfull : {selector : 'p,h1,h2,h3,h4,h5,h6,td,th,div,ul,ol,li,table,img', classes : 'monsti--htmlarea-align-full'},
      },
      content_css : ["/static/css/common.css","/site-static/css/site.css"],
    });
  });
})();

monsti.initAutoName = function() {
  var title = document.getElementById("Fields.core.Title");
  var name = document.getElementById("Name");
  if (!(title && name)) {
    return;
  }
  title.parentNode.parentNode.insertBefore(title.parentNode, name.parentNode);
  if (name.value != "") {
    return;
  }
  var active = true;
  var re1 = / /g;
  var re2 = /[^a-zA-Z0-9.-]/g;
  title.addEventListener("input", function(event) {
    if (!active) {
      return;
    }
    name.value = title.value.replace(re1, "-").replace(re2, "").toLowerCase();
  },false);
  title.addEventListener("change", function(event) {
    active = false;
  },false);
  name.addEventListener("change", function(event) {
    if (name.value == "") {
      active = true;
    }
  },false);
}

monsti.initEdit = function() {
  monsti.initAutoName();
}
