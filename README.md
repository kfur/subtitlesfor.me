# subtitlesfor.me 

Basic MVC Web App that convert audio to subtitles by using speech to text from IBM Watson API.
Audio can be provided from almost any video and audio formats except aac(due to IBM Watson API limitations)
thanks to ffmpeg.

Audio are converted to SRT caption/subtitle format from Watson STT generated JSON.

##Usage
To use this app you must have credentials from IBM Speech to Text service.
You can grab it on your IBM Cloud profile page.
 
Add your credentials to `config/config.json`:

~~~ json
{
...
    "SpeechToTextOptions": {
        "URL": "url to your speech to text service location",
        "IAMApiKey": "your key here"
    },
...
}
~~~

## Feedback

All feedback is welcome. Let me know if you have any suggestions, questions, or criticisms. 
If something is not idiomatic to Go, please let me know know so we can make it better.