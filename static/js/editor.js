// AddEditor transforms the given textarea into an editable for Aloha.
AddEditor = function(obj) {
    // TODO This won't work for multiple textareas.
    var $textarea = obj;
    $textarea.hide();
    var content = $textarea.val();
    $textarea.before("<div class=\"editable\">" + content + "</div>");
    editable = $('.editable')
    Aloha.jQuery(editable).aloha();
    $('.editor-submit').click(function() {
        editable.mahalo();
        $textarea.val(editable.html());
    });
}

$(document).ready(function () {
    Aloha.ready( function() {
        $('textarea.editor').each(function() { AddEditor($(this)); });
    });
});
