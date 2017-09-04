$(document).ready(function () {
    $('.tag').click(function () {
        var tag = $(this);
        if (tag.hasClass('badge-light')) {
            tag.removeClass('badge-light').addClass('badge-info').html('✔ ' + tag.html());
        } else if (tag.hasClass('badge-info')) {
            var html = 
            tag.removeClass('badge-info').addClass('badge-light').html(tag.html().replace('✔ ', ''));
        }
    });
});
/*
$(document).on('click', '#dropdownMenuButton span a', function (e) {
    console.log('PASO');
    console.log(e.which);
});
*/