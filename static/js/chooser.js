// Enable list action order dragging
(function() {
  var args = top.tinymce.activeEditor.windowManager.getParams();
  win = (args.window);
  input = (args.input);

  var selectButtons = document.querySelectorAll(".chooser-select");
  for (var i = 0; i < selectButtons.length; i++) {
    selectButtons[i].addEventListener("click", function(event) {
      var path = event.target.getAttribute("data-path");
      win.document.getElementById(input).value = path;
      top.tinymce.activeEditor.windowManager.close();
      return false;
    }, false);
  }
})();
