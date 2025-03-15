import { HttpClient } from '@angular/common/http';
import { Component } from '@angular/core';
import { MatListModule } from '@angular/material/list';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatCardModule } from '@angular/material/card';
import {MatSelectModule} from '@angular/material/select';
import {MatInputModule} from '@angular/material/input';
import {MatFormFieldModule} from '@angular/material/form-field';
import { MatButtonModule } from '@angular/material/button';
import { FormsModule } from '@angular/forms';
import { MatIconModule} from '@angular/material/icon';
import { RouterLink, RouterLinkActive } from '@angular/router';

interface Topic {
  name: string;
  description: string;
  id: number;
}

@Component({
  selector: 'app-topics',
  imports: [MatListModule, MatTooltipModule, MatCardModule, MatInputModule, MatSelectModule, MatFormFieldModule, MatButtonModule, FormsModule, RouterLink, RouterLinkActive, MatIconModule],
  templateUrl: './topics.component.html',
  styleUrl: './topics.component.scss'
})
export class TopicsComponent {
  topics: Topic[] = [];
  newTopicName = "";
  newTopicDescription = "";

  constructor(private http: HttpClient) {
    http.get<Topic[]>('/api/topic').subscribe((topics) => {
      console.log(topics);
      this.topics = topics;
      if (this.topics == null) {
        this.topics = [];
      }
    });
  }

  createTopic() {
    console.log("Creating topic", this.newTopicName, this.newTopicDescription);
    this.http.post<Topic>('/api/topic', {
      name: this.newTopicName,
      description: this.newTopicDescription
    }).subscribe((topic) => {
      this.topics.push(topic);
    });
  }

  deleteTopic(topic: Topic) {
    console.log("Deleting topic", topic);
    this.http.delete('/api/topic/'+topic.id).subscribe(() => {
      this.topics = this.topics.filter((t) => t != topic);
    });
  }
}
