// Integrate file chooser with CKEditor
(function() {
  // Add action for select file choosers' buttons
  var selectButtons = document.querySelectorAll(".chooser-select");
  for (var i = 0; i < selectButtons.length; i++) {
    selectButtons[i].addEventListener("click", function(event) {
      var path = event.target.getAttribute("data-path");
      var funcNum = monsti.getQueryParam('CKEditorFuncNum');
      window.opener.CKEDITOR.tools.callFunction(funcNum, path);
      window.close();
      return false;
    }, false);
  }

  // Add CKEditor parameters to URLs
  var links = document.querySelectorAll("a");
  for (var i = 0; i < links.length; i++) {
    links[i].href = links[i].href +
      "&CKEditor=" + monsti.getQueryParam('CKEditor') +
      "&CKEditorFuncNum=" + monsti.getQueryParam('CKEditorFuncNum') +
      "&langCode=" + monsti.getQueryParam('langCode');
  }
})();
