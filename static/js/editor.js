"use strict";

var monsti = monsti || {};

(function() {
  var browserPrefix = window.location.toString().replace(
    "@@edit", "@@chooser?type=");
  monsti.CKEditorConfig = {
    filebrowserBrowseUrl: browserPrefix,
    filebrowserImageBrowseUrl: browserPrefix + 'image',
    uiColor: '#E0D4C7',
    defaultLanguage: monsti.session.locale,
    removePlugins: 'maximize,stylescombo',
    removeButtons: 'SpecialChar',
    extraPlugins: 'autogrow,image2,showblocks',
    autoGrow_minHeight: 250,
    autoGrow_maxHeight: 600,
  };
})();

// addCKEditor adds an CKEditor for the textarea with the given id.
monsti.addCKEditor = function(elementId) {
  CKEDITOR.replace(elementId, monsti.CKEditorConfig);
}

// initAutoName initializes automatic node name filling.
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

// initEdit initializes edit action views.
monsti.initEdit = function() {
  monsti.initAutoName();
}

// getQueryParam returns the given parameter of the window's query.
monsti.getQueryParam = function(name) {
  var reParam = new RegExp('(?:[\?&]|&)' + name + '=([^&]+)', 'i') ;
  var match = window.location.search.match(reParam);
  if (match && match.length > 1) {
    return match[1];
  } else {
    return null;
  }
}
