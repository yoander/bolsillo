$(document).ready(function () {
    $('.tag').click(function () {
        var tag = $(this);
        //var elem = { "id": tag.attr('id'), "tag": tag.text() };
        var id = tag.attr('id')
        var input = $('input#tags');
        var tags = JSON.parse(input.val());
        if (tag.hasClass('badge-light')) {
            tags.push(id);
            tag.removeClass('badge-light').addClass('badge-info').html('✔ ' + tag.html());
        } else if (tag.hasClass('badge-info')) {
            tags.splice(tags.indexOf(id, 1));
            tag.removeClass('badge-info').addClass('badge-light').html(tag.html().replace('✔ ', ''));
        }
        tags = input.val(JSON.stringify(tags));
    });
});
/*
$(document).on('click', '#dropdownMenuButton span a', function (e) {
    console.log('PASO');
    console.log(e.which);
});
*/