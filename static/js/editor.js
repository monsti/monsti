AddEditor = function(obj) {
    var $textarea = obj;
    $textarea.before("<div id=\"epiceditor\"></div>");
    $textarea.hide();

    var editor = new EpicEditor({
        basePath: '/static/epiceditor',
        theme: {
            base:'/themes/base/epiceditor.css',
            preview:'/themes/preview/preview-light.css',
            editor:'/themes/editor/epic-light.css'
        }
    });
    editor.on('load', function (file) {
      $textarea.val(file.content);
    });
    editor.on('update', function (file) {
      $textarea.val(file.content);
    });
    editor.load();
}

$(document).ready(function () {
    $('textarea.editor').each(function() { AddEditor($(this)); });
});
