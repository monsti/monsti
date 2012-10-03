// ToggleHelp shows or hides the markup help.
ToggleHelp = function() {
    help = $('#editor-help');
    if(help.css('display') == "none") {
        help.load('/static/html/editor_help.de.html');
        help.slideDown();
    } else {
        help.slideUp();
    }
}

// AddEditor transforms the given textarea to a fancy javascript markup editor.
AddEditor = function(obj) {
    var $textarea = obj;
    $textarea.before("<div id=\"epiceditor\"></div>");
    $textarea.hide();
    var editor = new EpicEditor({
        basePath: '/static/epiceditor',
        clientSideStorage: false,
        theme: {
            base:'/themes/base/epiceditor.css',
            preview:'/themes/preview/preview-light.css',
            editor:'/themes/editor/epic-light.css'
        }
    });
    editor.on('load', function (file) {
        editor.importFile('epiceditor',
            $textarea.val());
    });
    editor.on('update', function (file) {
      $textarea.val(file.content);
    });
    editor.load();
    var help = '<div id="editor-help" style="display:none;">Loading...</div>'
    $('#epiceditor').append(help);
    var button = '<a id="btn-editor-help" href="#" title="Help"><img src="/static/img/icons/silk/help.png" alt="Help"/></a>';
    $('#epiceditor').before(button);
    $('#btn-editor-help').click(ToggleHelp);
}

$(document).ready(function () {
    $('textarea.editor').each(function() { AddEditor($(this)); });
});
