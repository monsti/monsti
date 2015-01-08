(function() {
  var monstiFileBrowser = function(field_name, url, type, win) {
    var cmsURL = window.location.toString();    // script URL - use an absolute path!
    cmsURL = cmsURL.replace("@@edit", "@@browse");
//    alert("Field_Name: " + field_name + " URL: " + url + " Type: " + type + " Win: " + win + " CMSURL: " + cmsURL); // debug/testing

    tinyMCE.activeEditor.windowManager.open({
      file : cmsURL,
      title : 'File Browser',
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
      file_browser_callback: monstiFileBrowser,
    });
  });
})();
