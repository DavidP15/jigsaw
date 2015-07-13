# jigsaw
A simple application that creates jigsaw pieces from an image given template images.

#Usage
You must include a config file. The file should be in json format. A sample is given in this project. The config file should look similar to this:
```
{
  "fullImageLocation": "image file location", /*relative to the image path given as a command line flag */
  "templateLocation": "templates", /* the folder that contains your template images */
  "pieces": [
    {
      "fileLocation": "piece1.png",
      "pieceLocationX": 0,
      "pieceLocationY": 0
    },
    ...
    {
      "fileLocation": "piece10.png",
      "pieceLocationX": 10, /* the image column*/
      "pieceLocationY": 0 /* the image row */
    }
  ],
  "pieceWidth": 43, /*width and height without any overflow. */
  "pieceHeight": 43,
  "pieceRows": 10, 
  "pieceColumns": 15,
  "pieceOverflow": 15,
  "TemplateOff": 0, /*the color for space that the piece does not occupy */
  "TemplateOn": 255 /*the color for space the piece does occupy */
}
```
There are two command line flags. They are init and images. init is the path to your initialization file. images is the path to your images folder.

You use like this:
```
jigsaw -init=path/to/your/init/file -images=path/to/your/images/folder
```
