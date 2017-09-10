$(document).ready(function () {
    $('.tag').click(function () {
        var tag = $(this);
        var elem = { "id": tag.attr('id'), "tag": tag.text() };
        var input = $('input#tags');
        var tags = JSON.parse(input.val());
        if (tag.hasClass('badge-light')) {
            tags.push(elem);
            tag.removeClass('badge-light').addClass('badge-info').html('✔ ' + tag.html());
        } else if (tag.hasClass('badge-info')) {
            console.log(tags.splice(tags.indexOf(elem, 1)));
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